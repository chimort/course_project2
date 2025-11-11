package service

import (
	"context"
	"fmt"

	"github.com/chimort/course_project2/api/proto/sharedpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/chimort/course_project2/iternal/middleware"
	"github.com/chimort/course_project2/iternal/user/converter"
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
		Username:  req.User.Username,
		FirstName: req.User.FirstName,
		LastName: req.User.LastName,
		Email:     req.User.Email,
		Password:  req.User.Password,
		Age:       int(req.User.Age),
		Gender:    req.User.Gender,
		Languages: converter.FromPbLanguages(req.User.Languages),
		Interests: converter.FromPbInterests(req.User.Interests),
	}

	fmt.Printf("DEBUG CreateUser: langs=%+v, ints=%+v\n", req.User.Languages, user.Interests)
	if err := s.service.CreateUser(ctx, user); err != nil {
		return &userpb.CreateUserResponse{Response: "failed"}, err
	}
	return &userpb.CreateUserResponse{Response: "success"}, nil
}

func (s *UserServer) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	username := req.Username

	user, err := s.service.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	userPb := &sharedpb.User{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName: user.LastName,
		Email:     user.Email,
		Password:  user.Password,
		Languages: converter.ToPbLanguages(user.Languages),
		Interests: converter.ToPbInterests(user.Interests),
		Age:       int32(user.Age),
		Gender:    user.Gender,
	}

	return &userpb.GetUserResponse{User: userPb}, nil
}

func (s *UserServer) GetProfile(ctx context.Context, req *userpb.GetProfileRequest) (*userpb.GetProfileResponse, error) {
    username, ok := ctx.Value(middleware.UsernameKey).(string)
    if !ok || username == "" {
        return nil, fmt.Errorf("unauthorized")
    }

    user, err := s.service.GetUser(ctx, username)
    if err != nil {
        return nil, err
    }

    return &userpb.GetProfileResponse{
        User: &sharedpb.User{
            Username:  user.Username,
            FirstName: user.FirstName,
            LastName:  user.LastName,
            Email:     user.Email,
			Password:   "",
            Age:       int32(user.Age),
            Gender:    user.Gender,
            Languages: converter.ToPbLanguages(user.Languages),
            Interests: converter.ToPbInterests(user.Interests),
        },
    }, nil
}
