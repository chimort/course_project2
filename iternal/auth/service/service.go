package service

import (
	"context"
	"fmt"
	"log"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
)

type AuthService struct {
	userClient userpb.UserServiceClient
}

func (s *AuthService) NewAuthServer(service *AuthService) any {
	panic("unimplemented")
}

func NewAuthService(userClient userpb.UserServiceClient) *AuthService {
	return &AuthService{userClient: userClient}
}

// Регистрация пользователя
func (s *AuthService) RegisterUser(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	user := req.GetUser()
	_, err := s.userClient.CreateUser(ctx, &userpb.CreateUserRequest{User: user})
	if err != nil {
		return &authpb.RegisterResponse{Status: "failed to register user"}, err
	}
	log.Printf("✅ User registered via user-service: %s", user.Username)
	return &authpb.RegisterResponse{Status: "registration successful"}, nil
}

// Логин пользователя
func (s *AuthService) LoginUser(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	// Мок, пока нет реальной базы
	if req.Username == "admin" && req.Password == "123" {
		return &authpb.LoginResponse{Token: "fake-jwt-token-for-admin"}, nil
	}
	return nil, fmt.Errorf("invalid credentials")
}
