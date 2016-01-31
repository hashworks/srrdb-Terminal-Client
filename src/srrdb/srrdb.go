package srrdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
)

const srrdbURL = "http://www.srrdb.com"

type SearchResponse struct {
	Results     []SearchResult `json:"results"`
	ResultCount string         `json:"resultsCount"`
	Warnings    []string       `json:"warnings"`
	Query       []string       `json:"query"`
}

type SearchResult struct {
	Dirname        string `json:"release"`
	DateResponse   string `json:"date"` // f.e. 2014-06-16 17:35:26
	HasNFOResponse string `json:"hasNFO"`
	HasSRSResponse string `json:"hasSRS"`
}

type UploadResult struct {
	Files []UploadedFile `json:"files"`
}

type UploadedFile struct {
	Dirname string `json:"name"`
	Color   int    `json:"color"`
	Message string `json:"message"`
}

func (r *SearchResult) HasNFO() bool {
	if r.HasNFOResponse == "yes" {
		return true
	}
	return false
}

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
	fmt.Println(string(bytes))
	err = json.Unmarshal(bytes, &srrDBResponse)
	fmt.Println(srrDBResponse)
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
	loginSucessfull := false
	for _, c := range response.Cookies() {
		if c.Name == "uid" {
			loginSucessfull = true
			break
		}
	}
	if loginSucessfull == false {
		return &cookiejar.Jar{}, errors.New("Wrong authentication?")
	}
	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return &cookiejar.Jar{}, err
	}
	jar.SetCookies(&url.URL{Scheme: "http", Host: "www.srrdb.com", Path: "/"}, response.Cookies())
	return jar, nil
}

// Upload will upload one or more SRR files to the srrdb.
// You can provide a login with a cookie jar, see NewLoginCookieJar().
func Upload(fps []string, jar *cookiejar.Jar) (UploadResult, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for _, fp := range fps {
		f, err := os.Open(fp)
		if err != nil {
			return UploadResult{}, err
		}
		fw, err := w.CreateFormFile("files[]", fp)
		if err != nil {
			return UploadResult{}, err
		}
		if _, err = io.Copy(fw, f); err != nil {
			return UploadResult{}, err
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", srrdbURL+"/release/upload", &b)
	if err != nil {
		return UploadResult{}, err
	}

	req.Header.Add("Content-Type", w.FormDataContentType())
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	client := &http.Client{Jar: jar}
	response, err := client.Do(req)
	if err != nil {
		return UploadResult{}, err
	}
	if response.StatusCode != 200 {
		return UploadResult{}, errors.New("Unexpected return code " + strconv.Itoa(response.StatusCode) + ".")
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return UploadResult{}, err
	}
	var uploadResult UploadResult
	err = json.Unmarshal(bytes, &uploadResult)
	return uploadResult, err
}
