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
		LastName:  req.User.LastName,
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
	user, err := s.service.GetUser(ctx, req.Username)
	if err != nil {
		return nil, err
	}

	return &userpb.GetUserResponse{User: buildUserPb(user, true)}, nil
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

	return &userpb.GetProfileResponse{User: buildUserPb(user, false)}, nil
}

func (s *UserServer) UpdateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*userpb.UpdateProfileResponse, error) {
	username, ok := ctx.Value(middleware.UsernameKey).(string)
	if !ok || username == "" {
		return nil, fmt.Errorf("unauthorized")
	}

	user := req.User
	updates := make(map[string]interface{})

	if user.FirstName != "" {
		updates["first_name"] = user.FirstName
	}
	if user.LastName != "" {
		updates["last_name"] = user.LastName
	}
	if user.Age != 0 {
		updates["age"] = int(user.Age)
	}
	if len(user.Languages) > 0 {
		updates["languages"] = converter.FromPbLanguages(user.Languages)
	}
	if len(user.Interests) > 0 {
		updates["interests"] = converter.FromPbInterests(user.Interests)
	}

	err := s.service.UpdateProfile(ctx, username, updates)
	if err != nil {
		return &userpb.UpdateProfileResponse{Status: "failed"}, err
	}

	return &userpb.UpdateProfileResponse{Status: "success"}, nil
}

func buildUserPb(user *models.User, includePassword bool) *sharedpb.User {
	pw := ""
	if includePassword {
		pw = user.Password
	}

	return &sharedpb.User{
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Password:  pw,
		Age:       int32(user.Age),
		Gender:    user.Gender,
		Languages: converter.ToPbLanguages(user.Languages),
		Interests: converter.ToPbInterests(user.Interests),
	}
}
