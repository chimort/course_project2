package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/gateway/handlers"
	"github.com/chimort/course_project2/iternal/gateway/middleware"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	ctx := context.Background()
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if err := authpb.RegisterRegisterServiceHandlerFromEndpoint(ctx, mux, "auth-service:50052", opts); err != nil {
		log.Error("failed to register auth gateway", "error", err)
	}

	conn, err := grpc.NewClient("user-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to new grpc client", "error", err)
		return
	}
	defer conn.Close()
	userClient := userpb.NewUserServiceClient(conn)

	e := echo.New()
	e.HideBanner = true

	userHandler := handlers.NewUserHandler(userClient, log)
	e.GET("/profile", userHandler.GetProfile, middleware.AuthMiddleware)

	e.File("/", "web/index.html")
	e.Static("/static", "web/static")

	e.Any("/*", echo.WrapHandler(mux))
	e.Start(":8080")
}
