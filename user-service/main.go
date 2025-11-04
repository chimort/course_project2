package main

import (
	"log"
	"net/http"

	"github.com/chimort/course_project2/handlers"
	"github.com/chimort/course_project2/store"
)

func main() {
	userStore := store.NewUserStore()
	userHandler := handlers.NewUserHandler(userStore)

	http.HandleFunc("/user/register", userHandler.RegisterUser)
	http.HandleFunc("/user/", userHandler.GetUser)

	log.Println("User service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}