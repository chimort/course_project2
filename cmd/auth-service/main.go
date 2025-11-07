package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type authServer struct {
    authpb.UnimplementedRegisterServiceServer
    client userpb.UserServiceClient
}

func newAuthServer(userClient userpb.UserServiceClient) *authServer {
    return &authServer{client: userClient}
}

func (s *authServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
    user := req.GetUser()

    // Делаем RPC вызов в user-service
    _, err := s.client.CreateUser(ctx, &userpb.CreateUserRequest{User:user})
    if err != nil {
        return &authpb.RegisterResponse{Status: "failed to register user"}, err
    }

    log.Printf("✅ User registered via user-service: %s", user.Username)
    return &authpb.RegisterResponse{Status: "registration successful"}, nil
}

func (s *authServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
    // В реальности — запрос к user-service, получение данных и валидация пароля
    // Пока просто мок
    if req.Username == "admin" && req.Password == "123" {
        return &authpb.LoginResponse{Token: "fake-jwt-token-for-admin"}, nil
    }
    return nil, fmt.Errorf("invalid credentials")
}

func main() {
    conn, err := grpc.NewClient("user-service:50051", 
    grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("❌ не удалось подключиться: %v", err)
    }
    defer conn.Close()
	userClient := userpb.NewUserServiceClient(conn)

    lis, err := net.Listen("tcp", ":50052")
    if err != nil {
        log.Fatalf("❌ failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    authpb.RegisterRegisterServiceServer(grpcServer, &authServer{client:userClient})

    fmt.Println("🚀 AuthService running on :50052 (connected to user-service:50051)")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}