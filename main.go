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

func ProxyRequestHandler(proxy *httputil.ReverseProxy, proxy2 *httputil.ReverseProxy, url string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conId := r.Header.Get("ConnectionId")

		if conId == "1" {
			proxy2.ServeHTTP(w, r)
		} else {
			proxy.ServeHTTP(w, r)
		}
	}
}

func createCollectionOfProxy(targetHosts []string) []*httputil.ReverseProxy {
	var proxies []*httputil.ReverseProxy
	for _, host := range targetHosts {
		proxy, err := NewProxy(host)
		if err != nil {
			log.Fatal(err)
		}
		proxies = append(proxies, proxy)
	}
	return proxies
}

func main() {
	var targetHosts []string
	targetHosts = append(targetHosts, "http://localhost:5233/test1")
	targetHosts = append(targetHosts, "http://localhost:5233/test2")
	var proxies = createCollectionOfProxy(targetHosts)

	http.HandleFunc("/", ProxyRequestHandler(proxies[0], proxies[1], ""))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
