package matching

import (
	"context"
	"log/slog"
	"time"

	"github.com/chimort/course_project2/api/proto/matchingpb"
	"github.com/redis/go-redis/v9"
)

const (
	queueKey     = "queue:matching"
	userQueueKey = "queue:user:" 
)

type MatchingService struct {
	redis *redis.Client
	log   *slog.Logger
}

func NewMatchingService(r *redis.Client, log *slog.Logger) *MatchingService {
	return &MatchingService{
		redis: r,
		log:   log.With("service", "matching_service"),
	}
}

func (s *MatchingService) JoinQueue(ctx context.Context, req *matchingpb.JoinQueueRequest) (*matchingpb.JoinQueueResponse, error) {
	if req == nil || req.Username == "" {
		return &matchingpb.JoinQueueResponse{Ok: false}, nil
	}

	_, err := s.redis.ZScore(ctx, queueKey, req.Username).Result()
	if err == nil {
		s.log.Warn("user already in queue", "username", req.Username)
		return &matchingpb.JoinQueueResponse{Ok: true}, nil
	} else if err != redis.Nil {
		s.log.Error("failed to check existing score", "error", err)
		return &matchingpb.JoinQueueResponse{Ok: false}, err
	}

	score := float64(time.Now().UnixMilli())

	z := redis.Z{
		Score:  score,
		Member: req.Username,
	}
	if err := s.redis.ZAdd(ctx, queueKey, z).Err(); err != nil {
		s.log.Error("failed to add user to queue", "error", err)
		return &matchingpb.JoinQueueResponse{Ok: false}, err
	}

	if err := s.redis.HSet(ctx, userQueueKey+req.Username, map[string]interface{}{
		"mode": req.Mode.String(),
	}).Err(); err != nil {
		s.log.Error("failed to store user params", "error", err)
		_ = s.redis.ZRem(ctx, queueKey, req.Username).Err()
		return &matchingpb.JoinQueueResponse{Ok: false}, err
	}

	s.log.Info("user joined queue", "username", req.Username, "mode", req.Mode.String(), "score", score)
	return &matchingpb.JoinQueueResponse{Ok: true}, nil
}

func (s *MatchingService) LeaveQueue(ctx context.Context, req *matchingpb.LeaveQueueRequest) (*matchingpb.LeaveQueueResponse, error) {
	if req == nil || req.Username == "" {
		return &matchingpb.LeaveQueueResponse{Ok: false}, nil
	}

	if err := s.redis.ZRem(ctx, queueKey, req.Username).Err(); err != nil {
		s.log.Error("failed to remove user from queue", "error", err)
		return &matchingpb.LeaveQueueResponse{Ok: false}, err
	}

	_ = s.redis.Del(ctx, userQueueKey+req.Username).Err()

	s.log.Info("user left queue", "username", req.Username)
	return &matchingpb.LeaveQueueResponse{Ok: true}, nil
}

func (s *MatchingService) ListQueue(ctx context.Context, req *matchingpb.ListQueueRequest) (*matchingpb.ListQueueResponse, error) {
	users, err := s.redis.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		s.log.Error("failed to list queue", "error", err)
		return nil, err
	}

	s.log.Info("queue listed", "count", len(users))
	return &matchingpb.ListQueueResponse{Usernames: users}, nil
}
