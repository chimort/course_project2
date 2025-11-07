package main

import (
	"fmt"
	"log"
	"net"

	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/user/service"
	"google.golang.org/grpc"
)

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("❌ failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()

    // создаём бизнес-логику
    userService := service.NewUserService()

    // регистрируем gRPC сервер с адаптером
    userpb.RegisterUserServiceServer(grpcServer, service.NewUserServer(userService))

    fmt.Println("🚀 UserService running on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
