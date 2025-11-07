package service

import (
    "context"

    "github.com/chimort/course_project2/api/proto/authpb"
)

type AuthServer struct {
    authpb.UnimplementedRegisterServiceServer
    service *AuthService // ссылка на бизнес-логику
}

// Register gRPC метод
func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
    return s.service.RegisterUser(ctx, req)
}

// Login gRPC метод
func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
    return s.service.LoginUser(ctx, req)
}

// Конструктор
func NewAuthServer(service *AuthService) *AuthServer {
    return &AuthServer{service: service}
}
