package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
)

type AuthService struct {
	userClient userpb.UserServiceClient
	log        *slog.Logger
}

func NewAuthService(userClient userpb.UserServiceClient, log *slog.Logger) *AuthService {
	return &AuthService{
		userClient: userClient,
		log:        log.With("service", "auth"),
	}
}

func (s *AuthService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	s.log.Info("register request received", "username", req.User.Username)

	user := req.GetUser()
	_, err := s.userClient.CreateUser(ctx, &userpb.CreateUserRequest{User: user})
	if err != nil {
		s.log.Error("failed to create user in user-service", "error", err)
		return &authpb.RegisterResponse{Status: "failed to register user"}, err
	}
	s.log.Info("user successfully registered", "username", req.User.Username)
	return &authpb.RegisterResponse{Status: "registration successful"}, nil
}

func (s *AuthService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	s.log.Info("login attempt", "username", req.Username)
	if req.Username == "admin" && req.Password == "123" {
		s.log.Info("login successful", "username", req.Username)
		return &authpb.LoginResponse{Token: "fake-jwt-token-for-admin"}, nil
	}
	s.log.Warn("login failed", "username", req.Username)
	return nil, fmt.Errorf("invalid credentials")
}
