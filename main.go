package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func NewProxy(targetHost string) (*httputil.ReverseProxy, error) {
	url, err := url.Parse(targetHost)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(url)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		proxy.ModifyResponse = modifyResponse(req)
	}

	proxy.ErrorHandler = errorHandler()
	return proxy, nil
}

func errorHandler() func(http.ResponseWriter, *http.Request, error) {
	return func(w http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Got error while modifying response: %v \n", err)
		return
	}
}

func modifyResponse(req *http.Request) func(*http.Response) error {
	authHeader := req.Header.Get("authy")
	if authHeader != "password" {
		return func(resp *http.Response) error {
			resp.Header.Add("Proxy", "Starship")
			resp.Body = ioutil.NopCloser(strings.NewReader(""))
			resp.StatusCode = 401
			return nil
		}
	} else {
		return nil
	}
}

var x = 1

func ProxyRequestHandler(proxy *httputil.ReverseProxy, proxy2 *httputil.ReverseProxy, url string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		x++
		if x%2 == 0 {
			proxy.ServeHTTP(w, r)
		} else {
			proxy2.ServeHTTP(w, r)
		}
	}
}

func main() {
	proxy, err := NewProxy("http://localhost:5233/test1")
	if err != nil {
		panic(err)
	}
	proxy2, err := NewProxy("http://localhost:5233/test2")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", ProxyRequestHandler(proxy, proxy2, ""))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
