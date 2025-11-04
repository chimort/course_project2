package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	userURL, _ := url.Parse("http://user-service:8080")
	matchingURL, _ := url.Parse("http://matching-service:8081")
	chatURL, _ := url.Parse("http://chat-service:8082")

	http.HandleFunc("/user/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = r.URL.Path[len("/user"):]
		httputil.NewSingleHostReverseProxy(userURL).ServeHTTP(w, r)
	})
	http.HandleFunc("/matching/", func(w http.ResponseWriter, r *http.Request) {
		httputil.NewSingleHostReverseProxy(matchingURL).ServeHTTP(w, r)
	})
	http.HandleFunc("/chat/", func(w http.ResponseWriter, r *http.Request) {
		httputil.NewSingleHostReverseProxy(chatURL).ServeHTTP(w, r)
	})

	port := ":8000"
	fmt.Println("API Gateway running on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
