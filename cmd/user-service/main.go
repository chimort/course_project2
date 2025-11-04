package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("user-service is alive"))
	})

	port := ":8080"
	fmt.Println("User service running on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}
