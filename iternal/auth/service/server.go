package service

import (
	"context"

	"github.com/chimort/course_project2/api/proto/authpb"
)

type AuthServer struct {
	authpb.UnimplementedRegisterServiceServer
	service *AuthService
}

func (s *AuthServer) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	return s.service.Register(ctx, req)
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	return s.service.Login(ctx, req)
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	return s.service.RefreshToken(ctx, req)
}

func NewAuthServer(service *AuthService) *AuthServer {
	return &AuthServer{service: service}
}
