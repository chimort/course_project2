package main

import (
	"log"
	"net/http"
	"os"

	"github.com/chimort/course_project2/iternal/matching"
	"github.com/redis/go-redis/v9"
)

func main() {
	redisHost := getenv("REDIS_HOST", "localhost:6379")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	svc := matching.NewService(rdb)

	mux := http.NewServeMux()
	mux.HandleFunc("/online/add", svc.HandleAddOnline)
	mux.HandleFunc("/online/list", svc.HandleListOnline)

	log.Println("matching-service started at :9000")
	http.ListenAndServe(":9000", mux)
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}
