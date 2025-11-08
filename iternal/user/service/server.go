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
		Language:  fromPbLanguages(req.User.Language),
		Interests: fromPbInterests(req.User.Interests),
	}

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
        Id:       user.ID,
        Username: user.Username,
        Password: user.Password,
        Language: toPbLanguages(user.Language),
        Interests: toPbInterests(user.Interests),
    }

    return &userpb.GetUserResponse{User: userPb}, nil
}

func mapSlice[T any, U any](in []T, f func(T) U) []U {
	out := make([]U, len(in))
	for i, v := range in {
		out[i] = f(v)
	}
	return out
}

func toPbLanguages(langs []models.Language) []*sharedpb.Language {
	return mapSlice(langs, func(l models.Language) *sharedpb.Language {
		return &sharedpb.Language{Name: string(l)}
	})
}

func toPbInterests(ints []models.Interests) []*sharedpb.Interests {
	return mapSlice(ints, func(i models.Interests) *sharedpb.Interests {
		return &sharedpb.Interests{Name: string(i)}
	})
}
func fromPbLanguages(langs []*sharedpb.Language) []models.Language {
	return mapSlice(langs, func(l *sharedpb.Language) models.Language {
		return models.Language(l.Name)
	})
}

func fromPbInterests(ints []*sharedpb.Interests) []models.Interests {
	return mapSlice(ints, func(i *sharedpb.Interests) models.Interests {
		return models.Interests(i.Name)
	})
}
