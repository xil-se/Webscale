package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
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
	ReplaceRobots bool
	Match         []Match
}

type Match struct {
	Search     string
	Replace    string
	DisableTmp bool
}

var (
	conf     Config
	confPath string
	vers     bool
	version  = "Not set"
)

const robotstxt = `User-agent: *
Disallow: /
`

func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	handling := "-"
	if strings.Contains(response.Header.Get("Content-Type"), "text/") {

		bin, readErr := ioutil.ReadAll(response.Body)
		if readErr != nil {
			log.Printf("Error reading: %s", readErr)
			return nil, readErr
		}

		body := bin
		var replace []byte
		for _, v := range conf.Match {
			if !v.DisableTmp {
				tmpl, err := template.New("replace").Parse(v.Replace)
				if err != nil {
					log.Printf("Could not parse template: %s with error: %s", v.Replace, err)
					continue
				}

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
			if response.Header.Get("Location") != "" {
				response.Header.Set("Location",
					string(regx.ReplaceAll([]byte(response.Header.Get("Location")),
						replace)))
			}
		}

		response.Header.Set("Content-Length", strconv.Itoa(len(body)))
		handling = "handled"
		response.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	log.Printf("%s %s %s %s %q %d %s\n", request.RemoteAddr, request.Host, request.Method, request.URL, request.Header.Get("User-Agent"), response.StatusCode, handling)
	return response, err
}
func robots(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, robotstxt)
}
func init() {
	flag.BoolVar(
		&vers,
		"version",
		false,
		"Print version")

	flag.StringVar(
		&confPath,
		"config",
		"./config",
		"Path to the config file")

	conf.ListenAddress = ":8080"
	conf.BackendScheme = "http"
	conf.BackendHost = "localhost"
	flag.Parse()
}

func main() {
	if vers {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}
	if _, err := toml.DecodeFile(confPath, &conf); err != nil {
		log.Fatal(err)
	}

	ts := &myTransport{}

	director := func(request *http.Request) {
		request.URL.Scheme = conf.BackendScheme
		request.URL.Host = conf.BackendHost
	}

	proxy := &httputil.ReverseProxy{Director: director, Transport: ts}
	if conf.ReplaceRobots {
		http.Handle("/robots.txt", http.HandlerFunc(robots))
	}
	http.Handle("/", proxy)

	fmt.Printf("Listening on %s\n", conf.ListenAddress)
	log.Fatal(http.ListenAndServe(conf.ListenAddress, nil))
}
