package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/chatpb"
	"github.com/chimort/course_project2/api/proto/matchingpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/gateway/handlers"
	gateway "github.com/chimort/course_project2/iternal/gateway/websocket"
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

	if err := matchingpb.RegisterMatchingServiceHandlerFromEndpoint(ctx, mux, "matching-service:50053", opts); err != nil {
		log.Error("failed to register matching gateway", "error", err)
		return
	}

	if err := chatpb.RegisterChatServiceHandlerFromEndpoint(ctx, mux, "chat-service:50054", opts); err != nil {
		log.Error("failed to register chat gateway", "error", err)
		return
	}

	// Create gRPC client for chat service
	chatConn, err := grpc.Dial("chat-service:50054", opts...)
	if err != nil {
		log.Error("failed to connect to chat service", "error", err)
		return
	}
	chatClient := chatpb.NewChatServiceClient(chatConn)

	e := echo.New()
	e.HideBanner = true
	e.File("/", "web/static/html/index.html")
	e.Static("/static", "web/static")

	hub := gateway.NewWSHub()
	wsHandler := handlers.NewWSHandler(hub, log, chatClient)

	e.GET("/ws", wsHandler.HandleWS)

	e.POST("iternal/ws/match-found", wsHandler.NotifyMatchFound)

	e.Any("/v1/*", echo.WrapHandler(mux))

	log.Info("🌐 API Gateway running on :8080")
	if err := e.Start(":8080"); err != nil {
		log.Error("server stopped with error", "error", err)
	}
}
