package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := authpb.RegisterRegisterServiceHandlerFromEndpoint(ctx, mux, "auth-service:50052", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	fmt.Println("🌐 API Gateway running on :8080")
	http.ListenAndServe(":8080", mux)
}
