package matching

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type Service struct {
	redis *redis.Client
	ctx   context.Context
}

func NewService(r *redis.Client) *Service {
	return &Service{
		redis: r,
		ctx:   context.Background(),
	}
}

func (s *Service) AddOnline(user string) error {
	return s.redis.SAdd(s.ctx, "online_users", user).Err()
}

func (s *Service) ListOnline() ([]string, error) {
	return s.redis.SMembers(s.ctx, "online_users").Result()
}
