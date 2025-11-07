package service

import (
	"context"

	"github.com/chimort/course_project2/api/proto/sharedpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/user/models"
)

type UserServer struct {
	userpb.UnimplementedUserServiceServer
	service *UserService
}

func NewUserServer(s *UserService) *UserServer {
	return &UserServer{service: s}
}

func (s *UserServer) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user := &models.User{
		ID:        req.User.Id,
		Username:  req.User.Username,
		Password:  req.User.Password,
		Language:  convertLanguages(req.User.Language),
		Interests: convertInterests(req.User.Interests),
	}

	if err := s.service.CreateUser(ctx, user); err != nil {
		return &userpb.CreateUserResponse{Response: "failed"}, err
	}
	return &userpb.CreateUserResponse{Response: "success"}, nil
}

func convertLanguages(items []*sharedpb.Language) []models.Language {
	langs := make([]models.Language, len(items))
	for i, l := range items {
		langs[i] = models.Language(l.Name)
	}
	return langs
}

func convertInterests(items []*sharedpb.Interests) []models.Interests {
	ints := make([]models.Interests, len(items))
	for i, in := range items {
		ints[i] = models.Interests(in.Name)
	}
	return ints
}
