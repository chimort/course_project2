package matching

import (
	"context"
	"log/slog"
	"math/rand"
	"sort"
	"time"

	"github.com/chimort/course_project2/api/proto/matchingpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/redis/go-redis/v9"
)

const (
	queueKey     = "queue:matching"
	userQueueKey = "queue:user:" // + username
)

// ----- сервис -----
type MatchingService struct {
	redis      *redis.Client
	userClient userpb.UserServiceClient
	log        *slog.Logger

	// Для тестов: позволяет подменить fetchProfiles
	fetchFunc func(ctx context.Context, usernames []string) (map[string]UserProfile, error)
}

func NewMatchingService(r *redis.Client, userClient userpb.UserServiceClient, log *slog.Logger) *MatchingService {
	return &MatchingService{
		redis:      r,
		userClient: userClient,
		log:        log.With("service", "matching_service"),
	}
}

// Устанавливаем fetchFunc (только для тестов)
func (s *MatchingService) SetFetchFunc(f func(ctx context.Context, usernames []string) (map[string]UserProfile, error)) {
	s.fetchFunc = f
}

// ----- очередь -----
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
	if err := s.redis.ZAdd(ctx, queueKey, redis.Z{Score: score, Member: req.Username}).Err(); err != nil {
		s.log.Error("failed to add user to queue", "error", err)
		return &matchingpb.JoinQueueResponse{Ok: false}, err
	}

	if err := s.redis.HSet(ctx, userQueueKey+req.Username, map[string]interface{}{"mode": req.Mode.String()}).Err(); err != nil {
		s.log.Error("failed to store user params", "error", err)
		_ = s.redis.ZRem(ctx, queueKey, req.Username).Err()
		return &matchingpb.JoinQueueResponse{Ok: false}, err
	}
	_ = s.redis.Expire(ctx, userQueueKey+req.Username, 120*time.Second).Err()

	s.log.Info("user joined queue", "username", req.Username, "mode", req.Mode.String())
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

// ----- fetch профилей -----
func (s *MatchingService) fetchProfiles(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
	if s.fetchFunc != nil {
		return s.fetchFunc(ctx, usernames)
	}

	if s.userClient == nil {
		// fallback: пустой профиль для продакшена
		out := make(map[string]UserProfile, len(usernames))
		for _, u := range usernames {
			out[u] = UserProfile{ID: u, Age: 30, Hobbies: []string{}, Language: "English"}
		}
		return out, nil
	}

	out := make(map[string]UserProfile, len(usernames))
	for _, uname := range usernames {
		resp, err := s.userClient.GetUser(ctx, &userpb.GetUserRequest{Username: uname})
		if err != nil {
			s.log.Warn("user service get failed", "username", uname, "err", err)
			continue
		}
		u := resp.User
		var lang string
		if len(u.Languages) > 0 {
			lang = u.Languages[0].Name
		}
		h := make([]string, 0, len(u.Interests))
		for _, it := range u.Interests {
			h = append(h, it.Name)
		}
		out[uname] = UserProfile{ID: u.Username, Age: int(u.Age), Hobbies: h, Language: lang}
		time.Sleep(5 * time.Millisecond)
	}
	return out, nil
}

// ----- fetch usernames из очереди -----
func (s *MatchingService) fetchQueueUsernames(ctx context.Context, limit int64) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.redis.ZRange(ctx, queueKey, 0, limit-1).Result()
}

// ----- FindBestMatch -----
func (s *MatchingService) FindBestMatch(ctx context.Context, username string, mode matchingpb.MatchMode) (string, error) {
	all, err := s.fetchQueueUsernames(ctx, 200)
	if err != nil {
		return "", err
	}

	var candidates []string
	for _, u := range all {
		if u != username {
			candidates = append(candidates, u)
		}
	}
	if len(candidates) == 0 {
		return "", nil
	}

	usernamesToFetch := append([]string{username}, candidates...)
	profiles, err := s.fetchProfiles(ctx, usernamesToFetch)
	if err != nil {
		return "", err
	}
	me, ok := profiles[username]
	if !ok {
		return "", nil
	}

	score, err := s.redis.ZScore(ctx, queueKey, username).Result()
	if err != nil {
		return "", err
	}
	waitSec := (float64(time.Now().UnixMilli()) - score) / 1000.0

	baseThreshold := 0.8
	minThreshold := 0.4
	decPerMin := 0.05
	threshold := baseThreshold - (waitSec/60.0)*decPerMin
	if threshold < minThreshold {
		threshold = minThreshold
	}

	type candidateScore struct {
		id    string
		total float64
		bd    ScoreBreakdown
	}
	var good []candidateScore
	var allScores []candidateScore

	for _, c := range candidates {
		p, ok := profiles[c]
		if !ok {
			continue
		}
		var prefs Preferences
		switch mode {
		case matchingpb.MatchMode_MATCH_MODE_LANGUAGE:
			prefs = Preferences{WeightLanguage: 0.6, WeightHobbies: 0.2, WeightAge: 0.2}
		case matchingpb.MatchMode_MATCH_MODE_INTEREST:
			prefs = Preferences{WeightHobbies: 0.6, WeightLanguage: 0.2, WeightAge: 0.2}
		default:
			prefs = Preferences{}
		}
		bd := CompatibilityDynamic(me, p, prefs)
		allScores = append(allScores, candidateScore{id: c, total: bd.Total, bd: bd})
		if bd.Total >= threshold {
			good = append(good, candidateScore{id: c, total: bd.Total, bd: bd})
		}
	}

	if len(good) >= 5 {
		sort.Slice(good, func(i, j int) bool { return good[i].total > good[j].total })
		top := good
		if len(top) > 5 {
			top = top[:5]
		}
		rand.Seed(time.Now().UnixNano())
		chosen := top[rand.Intn(len(top))].id
		_ = s.redis.ZRem(ctx, queueKey, username, chosen).Err()
		_ = s.redis.Del(ctx, userQueueKey+username, userQueueKey+chosen).Err()
		return chosen, nil
	}

	if len(good) > 0 {
		sort.Slice(good, func(i, j int) bool { return good[i].total > good[j].total })
		chosen := good[0].id
		_ = s.redis.ZRem(ctx, queueKey, username, chosen).Err()
		_ = s.redis.Del(ctx, userQueueKey+username, userQueueKey+chosen).Err()
		return chosen, nil
	}

	waitMin := waitSec / 60.0
	longWaitThresholdMin := 5.0
	if waitMin >= longWaitThresholdMin && len(allScores) > 0 {
		sort.Slice(allScores, func(i, j int) bool { return allScores[i].total > allScores[j].total })
		chosen := allScores[0].id
		_ = s.redis.ZRem(ctx, queueKey, username, chosen).Err()
		_ = s.redis.Del(ctx, userQueueKey+username, userQueueKey+chosen).Err()
		return chosen, nil
	}

	return "", nil
}