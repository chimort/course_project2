package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	log.Info("Starting API Gateway...")

	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := authpb.RegisterRegisterServiceHandlerFromEndpoint(ctx, mux, "auth-service:50052", opts); err != nil {
		log.Error("failed to register auth gateway", "error", err)
		return
	}

	if err := userpb.RegisterUserServiceHandlerFromEndpoint(ctx, mux, "user-service:50051", opts); err != nil {
		log.Error("failed to register user gateway", "error", err)
		return
	}

	e := echo.New()
	e.HideBanner = true
	e.File("/", "web/index.html")
	e.Static("/static", "web/static")

	e.Any("/v1/*", echo.WrapHandler(mux))

	log.Info("🌐 API Gateway running on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Error("server stopped with error", "error", err)
	}
}