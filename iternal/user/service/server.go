package service

import (
	"context"

	"github.com/chimort/course_project2/api/proto/userpb"
)

type userServer struct {
	userpb.UnimplementedUserServiceServer
	svc *UserService
}

func NewUserServer(svc *UserService) *userServer {
	return &userServer{svc: svc}
}

func (s *userServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	res := s.svc.CreateUser(req.GetUser())
	return &userpb.CreateUserResponse{Response: res}, nil
}
