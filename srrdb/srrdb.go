// Package srrdb allows to download & upload SRR files from srrdb.com and to access their search API.
package srrdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const srrdbURL = "http://www.srrdb.com"

// SearchResponse holds the results of a search.
type SearchResponse struct {
	Results     []SearchResult `json:"results"`
	ResultCount string         `json:"resultsCount"`
	Warnings    []string       `json:"warnings"`
	Query       []string       `json:"query"`
}

// SearchResult is the result of a search.
type SearchResult struct {
	Dirname        string `json:"release"`
	DateResponse   string `json:"date"` // f.e. 2014-06-16 17:35:26
	HasNFOResponse string `json:"hasNFO"`
	HasSRSResponse string `json:"hasSRS"`
}

// UploadResponse holds the results of a file upload.
type SRRUploadResponse struct {
	Files []SRRUploadedFile `json:"files"`
}

// UploadedFile is the result of a file upload.
type SRRUploadedFile struct {
	Dirname string `json:"name"`
	Color   int    `json:"color"`
	Message string `json:"message"`
}

// HasNFO will return if the search-result has a NFO file.
func (r *SearchResult) HasNFO() bool {
	if r.HasNFOResponse == "yes" {
		return true
	}
	return false
}

// HasSRS will return if the search-result has a SRS file.
func (r *SearchResult) HasSRS() bool {
	if r.HasSRSResponse == "yes" {
		return true
	}
	return false
}

// Search will request the srrdb search API with the provided query and return their response.
// For a list of available keywords see http://www.srrdb.com/help#keywords
func Search(query string) (SearchResponse, error) {
	client := http.DefaultClient
	sURL := srrdbURL + "/api/search/"
	queries := strings.Split(query, " ")
	for _, query := range queries {
		sURL = sURL + url.QueryEscape(query) + "/"
	}
	response, err := client.Get(sURL)
	if err != nil {
		return SearchResponse{}, err
	}
	if response.StatusCode != 200 {
		return SearchResponse{}, errors.New("Unexpected return code " + strconv.Itoa(response.StatusCode) + ".")
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return SearchResponse{}, err
	}
	var srrDBResponse SearchResponse
	err = json.Unmarshal(bytes, &srrDBResponse)
	return srrDBResponse, err
}

// Download will return a SRR file by dirname.
func Download(dirname string) ([]byte, error) {
	client := http.DefaultClient
	response, err := client.Get(srrdbURL + "/download/srr/" + url.QueryEscape(dirname))
	if err != nil {
		return []byte{}, err
	}
	if response.StatusCode != 200 {
		return []byte{}, errors.New("Unexpected return code " + strconv.Itoa(response.StatusCode) + ".")
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []byte{}, err
	}
	if string(body) == "The requested file does not exist." {
		return []byte{}, errors.New("Not found.")
	}
	return body, nil
}

// NewLoginCookieJar tries to login with the provided username and password and returns a cookie jar on success.
// You can store this jar, just take a look at the expiration date.
func NewLoginCookieJar(u, p string) (*cookiejar.Jar, error) {
	client := http.DefaultClient
	v := url.Values{}
	v.Set("username", u)
	v.Set("password", p)
	// srrdb.com will send a 302, to avoid Golang following that redirect and giving us wrong cookie data we'll return an error
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return errors.New("redirect")
	}
	response, err := client.PostForm(srrdbURL+"/account/login", v)
	if err != nil && err.Error() != "Post /: redirect" {
		return &cookiejar.Jar{}, err
	}
	if !containsValidLoginCookie(response.Cookies()) {
		return &cookiejar.Jar{}, errors.New("Wrong authentication?")
	}
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return &cookiejar.Jar{}, err
	}
	jar.SetCookies(srrdbURLStruct(), response.Cookies())
	return jar, nil
}

func containsValidLoginCookie(cookies []*http.Cookie) bool {
	for _, c := range cookies {
		if c.Name == "uid" {
			return true
		}
	}
	return false
}

func srrdbURLStruct() *url.URL {
	return &url.URL{Scheme: "http", Host: "www.srrdb.com", Path: "/"}
}

// UploadSRRs will upload one or more SRR files to the srrdb.
// You can provide a login with a cookie jar, see NewLoginCookieJar().
func UploadSRRs(fps []string, jar *cookiejar.Jar) (SRRUploadResponse, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for _, fp := range fps {
		f, err := os.Open(fp)
		if err != nil {
			return SRRUploadResponse{}, err
		}
		fw, err := w.CreateFormFile("files[]", filepath.Base(fp))
		if err != nil {
			return SRRUploadResponse{}, err
		}
		if _, err = io.Copy(fw, f); err != nil {
			return SRRUploadResponse{}, err
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", srrdbURL+"/release/upload", &b)
	if err != nil {
		return SRRUploadResponse{}, err
	}

	req.Header.Add("Content-Type", w.FormDataContentType())
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	client := &http.Client{Jar: jar}
	response, err := client.Do(req)
	if err != nil {
		return SRRUploadResponse{}, err
	}
	if response.StatusCode != 200 {
		return SRRUploadResponse{}, errors.New("Unexpected return code " + strconv.Itoa(response.StatusCode) + ".")
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return SRRUploadResponse{}, err
	}
	var uploadResult SRRUploadResponse
	err = json.Unmarshal(bytes, &uploadResult)
	return uploadResult, err
}

// UploadStoredFile will upload a stored file into a folder of a provided release to the srrdb.
// You must provide a valid login with a cookie jar, see NewLoginCookieJar().
func UploadStoredFile(fp, dirname, folder string, jar *cookiejar.Jar) (string, error) {
	if !containsValidLoginCookie(jar.Cookies(srrdbURLStruct())) {
		return "", errors.New("No login cookie found in provided cookie jar.")
	}
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	f, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	fw, err := w.CreateFormFile("file", filepath.Base(fp))
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, f); err != nil {
		return "", err
	}
	w.WriteField("folder", folder)
	w.WriteField("add", "")
	w.Close()

	req, err := http.NewRequest("POST", "http://www.srrdb.com/release/add/"+dirname, &b)
	if err != nil {
		return "", err
	}

	req.Header.Add("Content-Type", w.FormDataContentType())
	client := &http.Client{Jar: jar}
	response, err := client.Do(req)
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", errors.New("Unexpected return code " + strconv.Itoa(response.StatusCode) + ".")
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	res := regexp.MustCompile(`<div class="alert alert-.*>\r\s*([^<]*)`).FindStringSubmatch(string(bytes))
	if len(res) < 2 {
		return "", errors.New("Failed to parse upload result.")
	}
	return res[1], nil
}
