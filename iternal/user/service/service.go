package service

import "github.com/chimort/course_project2/api/proto/sharedpb"

// UserService хранит пользователей и методы бизнес-логики
type UserService struct {
    users map[string]*sharedpb.User
}

func NewUserService() *UserService {
    return &UserService{users: make(map[string]*sharedpb.User)}
}

// CreateUser содержит чистую бизнес-логику
func (s *UserService) CreateUser(user *sharedpb.User) string {
    if _, exists := s.users[user.Id]; exists {
        return "user already exists"
    }
    s.users[user.Id] = user
    return "user created successfully"
}
