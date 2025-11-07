package service

import (
	"context"
	"log/slog"

	"github.com/chimort/course_project2/iternal/user/models"
	"github.com/chimort/course_project2/iternal/user/repository"
)

type UserService struct {
	repo *repository.UserRepository
	log   *slog.Logger
}

func NewUserService(repo *repository.UserRepository, log *slog.Logger) *UserService {
	return &UserService{repo: repo, log: log}
}

func (s *UserService) CreateUser(ctx context.Context, u *models.User) error {
	s.log.Info("creating user", "username", u.Username)
	return s.repo.CreateUser(ctx, u)
}

func (s *UserService) GetUser(ctx context.Context, username string) (*models.User, error) {
	s.log.Info("fetching user", "username", username)
	return s.repo.GetUserByUsername(ctx, username)
}
