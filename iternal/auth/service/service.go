package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/chimort/course_project2/api/proto/authpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/auth/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	user.Id = uuid.New().String()
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

	resp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{Username: req.Username})
	if err != nil {
		s.log.Warn("login failed: user not found", "username", req.Username, "error", err)
		return nil, fmt.Errorf("invalid credentials")
	}
	user := resp.User
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.log.Warn("invalid password", "username", req.Username)
		return nil, fmt.Errorf("invalid credentials")
	}
	accessToken, err := token.GenerateJwt(user.Username)
	if err != nil {
		s.log.Error("failed to generate jwt", "error", err)
		return nil, fmt.Errorf("could not generate token")
	}
	refreshToken, err := token.GenerateRefreshToken(user.Username)
	if err != nil {
		s.log.Error("failed to generate jwt", "error", err)
		return nil, fmt.Errorf("could not generate token")
	}
	s.log.Info("login successful", "username", req.Username)
	return &authpb.LoginResponse{
		AccessToken: accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	claims, err := token.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userResp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{Username: claims.Username})
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	newAccessToken, err := token.GenerateJwt(userResp.User.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	newRefreshToken, err := token.GenerateRefreshToken(userResp.User.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

