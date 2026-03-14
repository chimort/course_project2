package main

import (
	"database/sql"
	"log/slog"
	"net"
	"os"

	_ "github.com/lib/pq"

	"github.com/chimort/course_project2/api/proto/chatpb"
	"github.com/chimort/course_project2/iternal/chat/repository"
	chat "github.com/chimort/course_project2/iternal/chat/service"
	"github.com/chimort/course_project2/iternal/pkg/logger"
	"google.golang.org/grpc"
)

func main() {
	logg := logger.NewLogger("chat-service", slog.LevelInfo)

	dsn := "host=db port=5432 user=postgres password=postgres dbname=users sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logg.Error("failed to connect db", "error", err)
		os.Exit(1)
	}

	chatRepo := repository.NewChatRepository(db)
	chatService := chat.NewChatService(chatRepo, logg)
	chatServer := chat.NewChatServer(chatService)

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		logg.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	chatpb.RegisterChatServiceServer(grpcServer, chatServer)

	logg.Info("chat-service started", "port", 50054)

	if err := grpcServer.Serve(lis); err != nil {
		logg.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
