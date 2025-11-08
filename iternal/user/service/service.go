package service

import (
	"context"
	"log/slog"

	"github.com/chimort/course_project2/iternal/user/models"
	"github.com/chimort/course_project2/iternal/user/repository"
	"golang.org/x/crypto/bcrypt"
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error("failed to hash password", "error", err)
		return err
	}
	u.Password = string(hashedPassword)
	
	return s.repo.CreateUser(ctx, u)
}

func (s *UserService) GetUser(ctx context.Context, username string) (*models.User, error) {
	s.log.Info("fetching user", "username", username)
	return s.repo.GetUserByUsername(ctx, username)
}
