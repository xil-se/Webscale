package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/BurntSushi/toml"
)

type myTransport struct {
	http.RoundTripper
}

type Config struct {
	ListenAddress string
	BackendScheme string
	BackendHost   string
	Match         []Match
}

type Match struct {
	Search     string
	Replace    string
	DisableTmp bool
}

var conf Config

func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	handling := "-"
	if val, ok := response.Header["Content-Type"]; ok {
		ctype := val[0]

		if strings.Contains(ctype, "text/") {

			bin, readErr := ioutil.ReadAll(response.Body)
			if readErr != nil {
				return nil, err
			}

			body := bin
			var replace []byte
			for _, v := range conf.Match {
				tmpl, err := template.New("replace").Parse(v.Replace)
				if err != nil {
					log.Printf("Could not parse template: %s with error: %s", v.Replace, err)
					continue
				}
				if !v.DisableTmp {
					buf := new(bytes.Buffer)
					err = tmpl.Execute(buf, request)
					if err != nil {
						log.Printf("Could not parse template: %s with error: %s", v.Replace, err)
						continue
					}
					replace = buf.Bytes()
				} else {
					replace = []byte(v.Replace)
				}

				regx := regexp.MustCompile(v.Search)
				body = regx.ReplaceAll(body, replace)
			}

			response.Header.Set("Content-Length", strconv.Itoa(len(body)))
			handling = "handled"
			response.Body = ioutil.NopCloser(bytes.NewReader(body))
		}
	}

	log.Printf("%s %s %s %s %q %d %s\n", request.RemoteAddr, request.Host, request.Method, request.URL, request.Header.Get("User-Agent"), response.StatusCode, handling)
	return response, err
}

func init() {
	conf.ListenAddress = ":8080"
	conf.BackendScheme = "http"
	conf.BackendHost = "localhost"
}

func main() {
	if _, err := toml.DecodeFile("config", &conf); err != nil {
		log.Fatal(err)
	}

	ts := &myTransport{}

	director := func(request *http.Request) {
		request.URL.Scheme = conf.BackendScheme
		request.URL.Host = conf.BackendHost
	}

	proxy := &httputil.ReverseProxy{Director: director, Transport: ts}
	server := http.Server{
		Addr:    conf.ListenAddress,
		Handler: proxy,
	}
	fmt.Printf("Listening on %s\n", conf.ListenAddress)
	server.ListenAndServe()
}
