package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/pkg/logger"
	"github.com/chimort/course_project2/iternal/user/service"
	"google.golang.org/grpc"
)

func main() {
	log := logger.NewLogger("user-service", slog.LevelInfo)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	userService := service.NewUserService(log)
	userpb.RegisterUserServiceServer(grpcServer, service.NewUserServer(userService))

	log.Info("UserService running", "addr", ":50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
