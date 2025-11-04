package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	userServiceURL, _ := url.Parse("http://user-service:8080")
	userProxy := httputil.NewSingleHostReverseProxy(userServiceURL)

	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/user")
		userProxy.ServeHTTP(w, r)
	})
	port := ":8000"
    fmt.Println("API Gateway running on", port)
    log.Fatal(http.ListenAndServe(port, nil))
}