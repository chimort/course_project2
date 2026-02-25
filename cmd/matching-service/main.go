package main

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/chimort/course_project2/api/proto/matchingpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	matching "github.com/chimort/course_project2/iternal/matching/service"
	"github.com/chimort/course_project2/iternal/pkg/logger"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logg := logger.NewLogger("matching-service", slog.LevelInfo)

	rdb := redis.NewClient(&redis.Options{
		Addr: "redis:6379",
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logg.Error("redis ping failed", "error", err)
		os.Exit(1)
	}

	userConn, err := grpc.Dial(
		"user-service:50051",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logg.Error("failed to dial user-service", "error", err)
		os.Exit(1)
	}
	defer userConn.Close()

	userClient := userpb.NewUserServiceClient(userConn)

	matchingService := matching.NewMatchingService(rdb, userClient, logg)
	matchingServer := matching.NewMatchingServer(matchingService)
	
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		logg.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	matchingpb.RegisterMatchingServiceServer(grpcServer, matchingServer)

	logg.Info("matching-service started", "port", 50053)

	if err := grpcServer.Serve(lis); err != nil {
		logg.Error("failed to serve", "error", err)
		os.Exit(1)
	}
}