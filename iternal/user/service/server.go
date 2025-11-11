package service

import (
	"context"
	"fmt"

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
		Username:  req.User.Username,
		FirstName: req.User.FirstName,
		LastName: req.User.LastName,
		Email:     req.User.Email,
		Password:  req.User.Password,
		Age:       int(req.User.Age),
		Gender:    req.User.Gender,
		Languages: fromPbLanguages(req.User.Languages),
		Interests: fromPbInterests(req.User.Interests),
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
		Languages: toPbLanguages(user.Languages),
		Interests: toPbInterests(user.Interests),
		Age:       int32(user.Age),
		Gender:    user.Gender,
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

func toPbLanguages(langs []models.UserLanguage) []*sharedpb.Language {
	return mapSlice(langs, func(l models.UserLanguage) *sharedpb.Language {
		return &sharedpb.Language{
			Name:  string(l.Language),
			Level: languageLevelToProto(l.Level),
		}
	})
}

func fromPbLanguages(langs []*sharedpb.Language) []models.UserLanguage {
	return mapSlice(langs, func(l *sharedpb.Language) models.UserLanguage {
		return models.UserLanguage{
			Language: models.Language(l.Name),
			Level:    languageLevelFromProto(l.Level),
		}
	})
}

func toPbInterests(ints []models.UserInterest) []*sharedpb.Interests {
	return mapSlice(ints, func(i models.UserInterest) *sharedpb.Interests {
		return &sharedpb.Interests{Name: string(i.Interest)}
	})
}

func fromPbInterests(ints []*sharedpb.Interests) []models.UserInterest {
	return mapSlice(ints, func(i *sharedpb.Interests) models.UserInterest {
		return models.UserInterest{Interest: models.Interests(i.Name)}
	})
}


func languageLevelToProto(level models.LanguageLevel) sharedpb.LanguageLevel {
	switch level {
	case "NATIVE":
		return sharedpb.LanguageLevel_NATIVE
	case "MEDIUM":
		return sharedpb.LanguageLevel_MEDIUM
	case "LOW":
		return sharedpb.LanguageLevel_LOW
	default:
		return sharedpb.LanguageLevel_LOW
	}
}

func languageLevelFromProto(level sharedpb.LanguageLevel) models.LanguageLevel {
	switch level {
	case sharedpb.LanguageLevel_NATIVE:
		return "NATIVE"
	case sharedpb.LanguageLevel_MEDIUM:
		return "MEDIUM"
	case sharedpb.LanguageLevel_LOW:
		return "LOW"
	default:
		return "LOW"
	}
}