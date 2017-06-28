package main

import (
	"flag"
	"fmt"
	"github.com/hashworks/go-srrdb-API/srrdb"
	"io/ioutil"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	// Set the following uppercase three with -ldflags "-X main.VERSION=v1.2.3 [...]"
	VERSION        string = "unknown"
	BUILD_COMMIT   string = "unknown"
	BUILD_DATE     string = "unknown"
	versionFlag    bool
	searchFlag     bool
	downloadFlag   bool
	extensionFlag  string
	stdoutFlag     bool
	prunePathsFlag bool
	uploadFlag     bool
	usernameFlag   string
	passwordFlag   string
	releaseFlag    string
	folderFlag     string
)

type storedFile struct {
	name string
	data []byte
}

func main() {
	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagSet.Usage = usage

	flagSet.BoolVar(&versionFlag, "version", false, "")
	flagSet.BoolVar(&versionFlag, "v", false, "")

	flagSet.BoolVar(&searchFlag, "search", false, "")
	flagSet.BoolVar(&searchFlag, "s", false, "")

	flagSet.BoolVar(&downloadFlag, "download", false, "")
	flagSet.BoolVar(&downloadFlag, "d", false, "")
	flagSet.StringVar(&extensionFlag, "extension", "", "")
	flagSet.StringVar(&extensionFlag, "e", "", "")
	flagSet.BoolVar(&stdoutFlag, "stdout", false, "")
	flagSet.BoolVar(&stdoutFlag, "o", false, "")
	flagSet.BoolVar(&prunePathsFlag, "prunePaths", false, "")

	flagSet.BoolVar(&uploadFlag, "upload", false, "")
	flagSet.BoolVar(&uploadFlag, "u", false, "")
	flagSet.StringVar(&usernameFlag, "username", "", "")
	flagSet.StringVar(&usernameFlag, "n", "", "")
	flagSet.StringVar(&passwordFlag, "password", "", "")
	flagSet.StringVar(&passwordFlag, "p", "", "")
	flagSet.StringVar(&releaseFlag, "release", "", "")
	flagSet.StringVar(&releaseFlag, "r", "", "")
	flagSet.StringVar(&folderFlag, "folder", "", "")
	flagSet.StringVar(&folderFlag, "f", "", "")

	flagSet.Parse(os.Args[1:])

	switch {
	case versionFlag:
		fmt.Println("srrdb.com Terminal Client")
		fmt.Println("https://github.com/hashworks/srrdb-Terminal-Client")
		fmt.Println("Version: " + VERSION)
		fmt.Println("Commit: " + BUILD_COMMIT)
		fmt.Println("Build date: " + BUILD_DATE)
		fmt.Println()
		fmt.Println("Published under the GNU General Public License v3.0.")
	case searchFlag:
		search(strings.Join(flagSet.Args(), " "))
	case downloadFlag:
		download(flagSet.Args(), extensionFlag, stdoutFlag, prunePathsFlag)
	case uploadFlag:
		if releaseFlag == "" {
			uploadSRRs(flagSet.Args(), usernameFlag, passwordFlag)
		} else {
			uploadStoredFiles(flagSet.Args(), releaseFlag, folderFlag, usernameFlag, passwordFlag)
		}
	default:
		flagSet.Usage()
	}
}

func usage() {
	fmt.Println("-v, --version")
	fmt.Println("	Shows the version and a few informations.")
	fmt.Println("")
	fmt.Println("-s, --search <query>[...]")
	fmt.Println("	Searches srrdb.com for releases.")
	fmt.Println("	For a list of available keywords see http://www.srrdb.com/help#keywords")
	fmt.Println("")
	fmt.Println("-d, --download <dirname>[...]")
	fmt.Println("	Download one or multiple SRR files from srrdb.com.")
	fmt.Println("	Options:")
	fmt.Println("	-e, --extension=<extension>")
	fmt.Println("		Saves only files with the specified extension from the SRR file.")
	fmt.Println("		You can prune file paths with --prunePaths.")
	fmt.Println("	-o, --stdout")
	fmt.Println("		Print file data to stdout instead of saving the file.")
	fmt.Println("")
	fmt.Println("-u, --upload <filename>[...]")
	fmt.Println("	Uploads one or multiple files to srrdb.com.")
	fmt.Println("	Options:")
	fmt.Println("	-n, --username=<username> and -p, --password=<password>")
	fmt.Println("		If you provide this it will post files using this account.")
	fmt.Println("	-r, --release=<dirname>")
	fmt.Println("		If you provide this it will post stored files to the specified release.")
	fmt.Println("		Note that you need a valid login for this.")
	fmt.Println("	-f, --folder=<folder>")
	fmt.Println("		Optional to --release, this will set the folder of the stored file.")
}

func search(query string) {
	response, err := srrdb.Search(query)
	if err != nil {
		fmt.Println("Failed to search for query: " + err.Error())
		os.Exit(1)
	}
	if response.ResultCount == "0" {
		fmt.Println("Nothing found!")
		os.Exit(1)
	}

	results := map[string]srrdb.SearchResult{}
	for _, r := range response.Results {
		results[r.DateResponse] = r
	}
	var keys []string
	for _, r := range results {
		keys = append(keys, r.DateResponse)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result := results[k]
		fmt.Print("[" + result.DateResponse + "] " + result.Dirname)
		if result.HasNFO() {
			fmt.Print(" [NFO]")
		}
		if result.HasSRS() {
			fmt.Print(" [SRS]")
		}
		fmt.Print("\n")
	}
}

func bytesToInt(b []byte) int {
	var r uint32
	for i := len(b) - 1; i >= 0; i-- {
		r |= uint32(b[i]) << uint32(i*8)
	}
	return int(r)
}

func isValidSRR(srr []byte) bool {
	if srr[0] != 0x69 || srr[1] != 0x69 || srr[2] != 0x69 {
		return true
	}
	return false
}

func extractStoredFiles(srr []byte) []storedFile {
	/*
		[SRR Stored File Block
		- HEAD_CRC: 0x6A6A                                  2 bytes
		- HEAD_TYPE: 0x6A                                   1 byte
		- HEAD_FLAGS:                                       2 bytes
			0x8000: must always be set to indicate the file size
		- HEAD_SIZE: limited to 65535 (0xFFFF) bytes        2 bytes
		- ADD_SIZE: the size of the stored file             4 bytes
		- NAME_SIZE: length of NAME string                  2 bytes
		- NAME: path and name of the stored file            NAME_SIZE bytes
		[Stored File Data]
		]
	*/
	var storedFiles []storedFile
	for i := 0; i < len(srr); i++ {
		if srr[i] == 0x6A && srr[i+1] == 0x6A && srr[i+1] == 0x6A {
			nameStart := i + 13
			nameSize := bytesToInt(srr[i+11 : nameStart])
			dataStart := nameStart + nameSize
			dataEnd := dataStart + bytesToInt(srr[i+7:i+11])
			storedFiles = append(storedFiles, storedFile{string(srr[nameStart:dataStart]), srr[dataStart:dataEnd]})
			i = dataEnd - 1
		}
	}
	return storedFiles
}

func saveFile(fp string, data []byte, pruneDir bool) {
	if pruneDir {
		fp = filepath.Base(fp)
	} else {
		os.MkdirAll(filepath.Dir(fp), os.ModePerm)
	}
	err := ioutil.WriteFile(fp, data, os.ModePerm)
	if err != nil {
		fmt.Println("Failed to save file to " + fp + ": " + err.Error())
	} else {
		fmt.Println("Saved file to " + fp + ".")
	}
}

func download(dirnames []string, extension string, toStdout, prunePaths bool) {
	if len(dirnames) == 0 {
		fmt.Println("You must provide at least one dirname.")
		os.Exit(1)
	}
	for _, dirname := range dirnames {
		srr, err := srrdb.Download(dirname)
		if err != nil {
			fmt.Println("Failed to download SRR file for " + dirname + ": " + err.Error())
		} else {
			if isValidSRR(srr) {
				fmt.Println("The downloaded file for " + dirname + " isn't a valid SRR file.")
			} else {
				extension = strings.ToLower(extension)
				if extension == "" || extension == "srr" {
					if toStdout {
						fmt.Print(string(srr))
					} else {
						saveFile(dirname+".srr", srr, prunePaths)
					}
				} else {
					storedFiles := extractStoredFiles(srr)
					fileFound := false
					for _, file := range storedFiles {
						if strings.ToLower(file.name[len(file.name)-len(extension):]) == extension {
							if toStdout {
								os.Stdout.Write(file.data)
							} else {
								saveFile(file.name, file.data, prunePaths)
							}
							fileFound = true
						}
					}
					if !fileFound {
						fmt.Println("Extension not found in SRR of " + dirname + ".")
					}
				}
			}
		}
	}
}

func uploadSRRs(fps []string, username, password string) {
	if len(fps) == 0 {
		fmt.Println("You must provide at least one file to upload.")
		os.Exit(1)
	}

	var (
		jar *cookiejar.Jar
		err error
	)

	if username != "" && password != "" {
		jar, err = srrdb.NewLoginCookieJar(username, password)
		if err != nil {
			fmt.Println("Failed to login: " + err.Error())
			os.Exit(1)
		}
	} else {
		jar, _ = cookiejar.New(&cookiejar.Options{})
	}

	response, err := srrdb.UploadSRRs(fps, jar)
	if err != nil {
		fmt.Println("Failed to upload SRR files: " + err.Error())
		os.Exit(1)
	}
	for _, file := range response.Files {
		messageLen := len(file.Message)
		if messageLen >= len(file.Dirname) && file.Dirname == file.Message[:len(file.Dirname)] {
			fmt.Println(file.Message)
		} else {
			if messageLen >= 3 && file.Message[:3] == " - " {
				fmt.Println(file.Dirname + file.Message)
			} else {
				fmt.Println(file.Dirname + " - " + file.Message)
			}
		}
	}
}

func uploadStoredFiles(fps []string, dirname, folder, username, password string) {
	if len(fps) == 0 {
		fmt.Println("You must provide at least one file to upload.")
		os.Exit(1)
	}
	if usernameFlag == "" || passwordFlag == "" {
		fmt.Println("You need to set your username and password to upload stored files.")
		os.Exit(1)
	}

	var (
		jar *cookiejar.Jar
		err error
	)

	jar, err = srrdb.NewLoginCookieJar(username, password)
	if err != nil {
		fmt.Println("Failed to login: " + err.Error())
		os.Exit(1)
	}

	for i := 0; i < len(fps); i++ {
		fp := fps[i]
		response, err := srrdb.UploadStoredFile(fp, dirname, folder, jar)
		fmt.Print(filepath.Base(fp) + ": ")
		if err != nil {
			fmt.Println("Failed to upload stored file - " + err.Error())
		} else {
			fmt.Println(response)
		}
	}
}
