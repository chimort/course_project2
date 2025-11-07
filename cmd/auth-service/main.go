package main

import (
	"fmt"
	"log"
	"net"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/auth/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    // Подключение к user-service
    conn, err := grpc.NewClient("user-service:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
    )
    if err != nil {
        log.Fatalf("❌ не удалось подключиться: %v", err)
    }
    defer conn.Close()
    
    userClient := userpb.NewUserServiceClient(conn)

    // Создаём бизнес-логику и gRPC сервер
    authService := service.NewAuthService(userClient)
    authSrv := service.NewAuthServer(authService)

    // Запуск gRPC сервера
    lis, err := net.Listen("tcp", ":50052")
    if err != nil {
        log.Fatalf("❌ failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    authpb.RegisterRegisterServiceServer(grpcServer, authSrv)

    fmt.Println("🚀 AuthService running on :50052 (connected to user-service:50051)")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
