package main

import (
	"log/slog"
	"net"
	"os"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/auth/service"
	"github.com/chimort/course_project2/iternal/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logg := logger.NewLogger("auth-service", slog.LevelInfo)

	conn, err := grpc.NewClient("user-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logg.Error("failed to connect to user-service", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	userClient := userpb.NewUserServiceClient(conn)
	authService := service.NewAuthService(userClient, logg)
	authSrv := service.NewAuthServer(authService)

	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		logg.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterRegisterServiceServer(grpcServer, authSrv)

	logg.Info("auth-service started", "port", 50052)
	if err := grpcServer.Serve(lis); err != nil {
		logg.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}
