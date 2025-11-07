package service

import (
	"log/slog"

	"github.com/chimort/course_project2/api/proto/sharedpb"
)

type UserService struct {
	users map[string]*sharedpb.User
	log   *slog.Logger
}

func NewUserService(log *slog.Logger) *UserService {
	return &UserService{users: make(map[string]*sharedpb.User), log: log.With("service", "user")}
}

func (s *UserService) CreateUser(user *sharedpb.User) string {
	if _, exists := s.users[user.Id]; exists {
		s.log.Warn("user already exists", "username", user.Username)
		return "user already exists"
	}
	s.users[user.Id] = user
	s.log.Info("user created", "username", user.Username)
	return "user created successfully"
}
