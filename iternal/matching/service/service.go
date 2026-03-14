package matching

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/chimort/course_project2/api/proto/chatpb"
	"github.com/chimort/course_project2/api/proto/matchingpb"
	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/metadata"
)

const (
	queueKey     = "queue:matching"
	userQueueKey = "queue:user:" // + username
)

// MatchingResult describes result of matching
type MatchingResult struct {
	Username        string
	Reason          string
	CommonInterests []string
}

// MatchingService - main service
type MatchingService struct {
	redis      *redis.Client
	userClient userpb.UserServiceClient
	chatClient chatpb.ChatServiceClient
	log        *slog.Logger
	notifyURL  string

	// for tests: ability to override profile fetcher
	fetchFunc func(ctx context.Context, usernames []string) (map[string]UserProfile, error)

	cacheMu      sync.RWMutex
	profileCache map[string]cachedProfile
	cacheTTL     time.Duration
}

type cachedProfile struct {
	profile UserProfile
	expires time.Time
}

func NewMatchingService(r *redis.Client, userClient userpb.UserServiceClient,
	chatClient chatpb.ChatServiceClient, log *slog.Logger) *MatchingService {
	return &MatchingService{
		redis:        r,
		userClient:   userClient,
		chatClient:   chatClient,
		log:          log.With("service", "matching_service"),
		notifyURL:    "http://gateway:8080/iternal/ws/match-found",
		profileCache: make(map[string]cachedProfile),
		cacheTTL:     3 * time.Second,
	}
}

// SetFetchFunc used by tests to inject fake profiles
func (s *MatchingService) SetFetchFunc(f func(ctx context.Context, usernames []string) (map[string]UserProfile, error)) {
	s.fetchFunc = f
}

// JoinQueue - add user to sorted set with timestamp score
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

	if err := s.redis.HSet(ctx, userQueueKey+req.Username, map[string]interface{}{
		"mode":          req.Mode.String(),
		"language_mode": normalizeLanguageMode(req.LanguageMode).String(),
	}).Err(); err != nil {
		s.log.Error("failed to store user params", "error", err)
		_ = s.redis.ZRem(ctx, queueKey, req.Username).Err()
		return &matchingpb.JoinQueueResponse{Ok: false}, err
	}
	_ = s.redis.Expire(ctx, userQueueKey+req.Username, 120*time.Second).Err()

	s.log.Info("user joined queue", "username", req.Username, "mode", req.Mode.String())
	return &matchingpb.JoinQueueResponse{Ok: true}, nil
}

// LeaveQueue - remove user from queue
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

// ListQueue - get all usernames in queue
func (s *MatchingService) ListQueue(ctx context.Context, req *matchingpb.ListQueueRequest) (*matchingpb.ListQueueResponse, error) {
	users, err := s.redis.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		s.log.Error("failed to list queue", "error", err)
		return nil, err
	}
	s.log.Info("queue listed", "count", len(users))
	return &matchingpb.ListQueueResponse{Usernames: users}, nil
}

// fetchQueueUsernames - helper
func (s *MatchingService) fetchQueueUsernames(ctx context.Context, limit int64) ([]string, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.redis.ZRange(ctx, queueKey, 0, limit-1).Result()
}

// fetchProfiles - either calls user service or uses injected fetchFunc or returns simple defaults
func (s *MatchingService) fetchProfiles(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
	if s.fetchFunc != nil {
		return s.fetchFunc(ctx, usernames)
	}

	if s.userClient == nil {
		out := make(map[string]UserProfile, len(usernames))
		for _, u := range usernames {
			out[u] = UserProfile{
				ID:        u,
				Age:       30,
				Hobbies:   []string{},
				Languages: []LanguageSkill{{Name: "English", Level: "NATIVE"}},
			}
		}
		return out, nil
	}

	out := make(map[string]UserProfile, len(usernames))
	for _, uname := range usernames {
		if p, ok := s.getCachedProfile(uname); ok {
			out[uname] = p
			continue
		}

		iternalCtx := metadata.AppendToOutgoingContext(ctx, "iternal", "true")
		resp, err := s.userClient.GetUserForMatching(iternalCtx, &userpb.GetUserRequest{Username: uname})
		if err != nil {
			s.log.Warn("user service get failed", "username", uname, "err", err)
			continue
		}
		u := resp.User
		languages := make([]LanguageSkill, 0, len(u.Languages))
		for _, lang := range u.Languages {
			languages = append(languages, LanguageSkill{
				Name:  lang.Name,
				Level: lang.Level.String(),
			})
		}
		h := make([]string, 0, len(u.Interests))
		for _, it := range u.Interests {
			h = append(h, it.Name)
		}
		p := UserProfile{ID: u.Username, Age: int(u.Age), Hobbies: h, Languages: languages}
		out[uname] = p
		s.setCachedProfile(uname, p)
		// tiny throttle
		time.Sleep(5 * time.Millisecond)
	}
	return out, nil
}

func (s *MatchingService) getCachedProfile(username string) (UserProfile, bool) {
	s.cacheMu.RLock()
	item, ok := s.profileCache[username]
	s.cacheMu.RUnlock()
	if !ok {
		return UserProfile{}, false
	}

	if time.Now().After(item.expires) {
		s.cacheMu.Lock()
		delete(s.profileCache, username)
		s.cacheMu.Unlock()
		return UserProfile{}, false
	}

	return item.profile, true
}

func (s *MatchingService) setCachedProfile(username string, profile UserProfile) {
	s.cacheMu.Lock()
	s.profileCache[username] = cachedProfile{
		profile: profile,
		expires: time.Now().Add(s.cacheTTL),
	}
	s.cacheMu.Unlock()
}

// helper: intersect of hobbies (max 2 returned)
func intersect(a, b []string) []string {
	set := make(map[string]struct{}, len(a))
	for _, x := range a {
		set[strings.ToLower(x)] = struct{}{}
	}
	out := make([]string, 0, 2)
	for _, x := range b {
		if _, ok := set[strings.ToLower(x)]; ok {
			out = append(out, x)
			if len(out) >= 2 {
				break
			}
		}
	}
	return out
}

func hasInterest(p UserProfile, interest string) bool {
	for _, h := range p.Hobbies {
		if strings.EqualFold(h, interest) {
			return true
		}
	}
	return false
}

func interestTopicForMode(mode matchingpb.MatchMode) string {
	switch mode {
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE:
		return "movies"
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_MUSIC:
		return "music"
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_BOOKS:
		return "books"
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_SPORT:
		return "sport"
	default:
		return ""
	}
}

// FindBestMatch - core algorithm
// returns MatchingResult or nil if none found
func (s *MatchingService) FindBestMatch(ctx context.Context, username string, mode matchingpb.MatchMode) (*MatchingResult, error) {
	return s.findBestMatchWithSeen(ctx, username, mode, matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE, nil)
}

func pairKey(a, b string) string {
	if a < b {
		return a + "|" + b
	}
	return b + "|" + a
}

func (s *MatchingService) findBestMatchWithSeen(
	ctx context.Context,
	username string,
	mode matchingpb.MatchMode,
	languageMode matchingpb.LanguageMatchMode,
	seenPairs map[string]struct{},
) (*MatchingResult, error) {
	all, err := s.fetchQueueUsernames(ctx, 200)
	if err != nil {
		return nil, err
	}

	var candidates []string
	for _, u := range all {
		if u != username {
			candidates = append(candidates, u)
		}
	}
	if len(candidates) == 0 {
		return nil, nil
	}

	usernamesToFetch := append([]string{username}, candidates...)
	profiles, err := s.fetchProfiles(ctx, usernamesToFetch)
	if err != nil {
		return nil, err
	}

	me, ok := profiles[username]
	if !ok {
		return nil, nil
	}

	score, err := s.redis.ZScore(ctx, queueKey, username).Result()
	if err != nil {
		return nil, err
	}

	waitSec := (float64(time.Now().UnixMilli()) - score) / 1000.0

	baseThreshold := 0.8
	minThreshold := 0.4
	decPerMin := 0.06

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
	topic := interestTopicForMode(mode)

	for _, c := range candidates {
		if seenPairs != nil {
			key := pairKey(username, c)
			if _, ok := seenPairs[key]; ok {
				continue
			}
			seenPairs[key] = struct{}{}
		}

		p, ok := profiles[c]
		if !ok {
			continue
		}

		if topic != "" {
			if !hasInterest(me, topic) || !hasInterest(p, topic) {
				continue
			}
		}

		var prefs Preferences
		switch mode {
		case matchingpb.MatchMode_MATCH_MODE_LANGUAGE:
			prefs = Preferences{WeightLanguage: 0.7, WeightHobbies: 0.15, WeightAge: 0.15}
		case matchingpb.MatchMode_MATCH_MODE_INTEREST:
			prefs = Preferences{WeightHobbies: 0.6, WeightLanguage: 0.2, WeightAge: 0.2}
		case matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
			matchingpb.MatchMode_MATCH_MODE_DISCUSS_MUSIC,
			matchingpb.MatchMode_MATCH_MODE_DISCUSS_BOOKS,
			matchingpb.MatchMode_MATCH_MODE_DISCUSS_SPORT:
			prefs = Preferences{WeightLanguage: 0.5, WeightAge: 0.4, WeightHobbies: 0.1}
			prefs.HobbyTopic = topic
			prefs.HobbyTopicBoost = 0.35
		default:
			prefs = Preferences{}
		}

		bd := CompatibilityDynamic(me, p, prefs)
		if mode == matchingpb.MatchMode_MATCH_MODE_LANGUAGE && languageMode == matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_LEARNING_GOALS {
			wLang, _, _ := normalizeWeights(prefs)
			bd.RawLanguage = bestLanguageExchangeScore(me, p)
			bd.WeightedLang = bd.RawLanguage * wLang
			bd.Total = bd.WeightedLang + bd.WeightedHobby + bd.WeightedAge
		}

		allScores = append(allScores, candidateScore{id: c, total: bd.Total, bd: bd})

		if bd.Total >= threshold {
			good = append(good, candidateScore{id: c, total: bd.Total, bd: bd})
		}
	}

	removeFromQueue := func(u1, u2 string) {
		_ = s.redis.ZRem(ctx, queueKey, u1, u2).Err()
		_ = s.redis.Del(ctx, userQueueKey+u1, userQueueKey+u2).Err()
	}

	buildResult := func(chosen string, bd ScoreBreakdown) *MatchingResult {
		var reasonKey string

		if mode == matchingpb.MatchMode_MATCH_MODE_LANGUAGE {
			reasonKey = "language_exchange"
		} else if topic != "" {
			reasonKey = topic
		} else {
			max := bd.WeightedLang
			reasonKey = "language"

			if bd.WeightedHobby > max {
				max = bd.WeightedHobby
				reasonKey = "shared_interests"
			}
			if bd.WeightedAge > max {
				reasonKey = "similar_age"
			}
		}

		reason := "best match: " + reasonKey
		common := intersect(me.Hobbies, profiles[chosen].Hobbies)
		tags := make([]string, 0, 3)
		tags = append(tags, reasonKey)
		for _, c := range common {
			if len(tags) >= 3 {
				break
			}
			tags = append(tags, strings.ToLower(strings.TrimSpace(c)))
		}
		matchHint := tags[rand.Intn(len(tags))]

		removeFromQueue(username, chosen)

		if s.chatClient != nil {
			resp, err := s.chatClient.CreateChat(ctx, &chatpb.CreateChatRequest{
				User1: username,
				User2: chosen,
			})
			if err != nil {
				s.log.Error("failed to create chat",
					"user1", username,
					"user2", chosen,
					"error", err,
				)
			} else {
				_, err = s.chatClient.SetChatMatchTags(ctx, &chatpb.SetChatMatchTagsRequest{
					ChatId: resp.ChatId,
					Tags:   tags,
				})
				if err != nil {
					s.log.Warn("failed to set chat match tags", "chat_id", resp.ChatId, "error", err)
				}

				s.log.Info("chat created via chat-service",
					"chat_id", resp.ChatId,
					"user1", username,
					"user2", chosen,
				)
				s.notifyMatchFound(username, chosen, resp.ChatId, matchHint)
			}
		}

		return &MatchingResult{
			Username:        chosen,
			Reason:          reason,
			CommonInterests: common,
		}
	}

	// 1) many good
	if len(good) >= 5 {
		sort.Slice(good, func(i, j int) bool { return good[i].total > good[j].total })

		top := good
		if len(top) > 5 {
			top = top[:5]
		}

		src := rand.NewSource(time.Now().UnixNano())
		rng := rand.New(src)
		chosen := top[rng.Intn(len(top))]

		return buildResult(chosen.id, chosen.bd), nil
	}

	// 2) best good
	if len(good) > 0 {
		sort.Slice(good, func(i, j int) bool { return good[i].total > good[j].total })
		chosen := good[0]
		return buildResult(chosen.id, chosen.bd), nil
	}

	// 3) fallback long wait
	waitMin := waitSec / 60.0
	if waitMin >= 5.0 && len(allScores) > 0 {
		sort.Slice(allScores, func(i, j int) bool { return allScores[i].total > allScores[j].total })
		chosen := allScores[0]
		return buildResult(chosen.id, chosen.bd), nil
	}

	return nil, nil
}

func (s *MatchingService) StartMatcher(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	s.log.Info("matcher started")

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.log.Info("matcher stopped")
				return

			case <-ticker.C:
				s.runMatchCycle(ctx)
			}
		}
	}()
}

func (s *MatchingService) runMatchCycle(ctx context.Context) {
	users, err := s.fetchQueueUsernames(ctx, 50)
	if err != nil {
		s.log.Warn("failed to fetch queue usernames", "error", err)
		return
	}

	if len(users) < 2 {
		return
	}

	seenPairs := make(map[string]struct{}, len(users))

	for _, username := range users {
		mode, err := s.getUserMode(ctx, username)
		if err != nil {
			continue
		}
		languageMode, err := s.getUserLanguageMode(ctx, username)
		if err != nil {
			languageMode = matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE
		}

		_, err = s.findBestMatchWithSeen(ctx, username, mode, languageMode, seenPairs)
		if err != nil {
			s.log.Warn("FindBestMatch failed", "username", username, "error", err)
		}
	}
}

func (s *MatchingService) getUserMode(ctx context.Context, username string) (matchingpb.MatchMode, error) {
	val, err := s.redis.HGet(ctx, userQueueKey+username, "mode").Result()
	if err != nil {
		return matchingpb.MatchMode_MATCH_MODE_DEFAULT, err
	}

	switch val {
	case matchingpb.MatchMode_MATCH_MODE_LANGUAGE.String():
		return matchingpb.MatchMode_MATCH_MODE_LANGUAGE, nil
	case matchingpb.MatchMode_MATCH_MODE_INTEREST.String():
		return matchingpb.MatchMode_MATCH_MODE_INTEREST, nil
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE.String():
		return matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE, nil
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_MUSIC.String():
		return matchingpb.MatchMode_MATCH_MODE_DISCUSS_MUSIC, nil
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_BOOKS.String():
		return matchingpb.MatchMode_MATCH_MODE_DISCUSS_BOOKS, nil
	case matchingpb.MatchMode_MATCH_MODE_DISCUSS_SPORT.String():
		return matchingpb.MatchMode_MATCH_MODE_DISCUSS_SPORT, nil
	default:
		return matchingpb.MatchMode_MATCH_MODE_DEFAULT, nil
	}
}

func normalizeLanguageMode(mode matchingpb.LanguageMatchMode) matchingpb.LanguageMatchMode {
	if mode == matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_UNSPECIFIED {
		return matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE
	}
	return mode
}

func (s *MatchingService) getUserLanguageMode(ctx context.Context, username string) (matchingpb.LanguageMatchMode, error) {
	val, err := s.redis.HGet(ctx, userQueueKey+username, "language_mode").Result()
	if err != nil {
		return matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE, err
	}

	switch val {
	case matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_LEARNING_GOALS.String():
		return matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_LEARNING_GOALS, nil
	case matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE.String():
		return matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE, nil
	default:
		return matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE, nil
	}
}

func (s *MatchingService) notifyMatchFound(user1, user2, chatID, matchHint string) {
	if s.notifyURL == "" {
		return
	}

	body, err := json.Marshal(map[string]string{
		"user1":      user1,
		"user2":      user2,
		"chat_id":    chatID,
		"match_hint": matchHint,
	})
	if err != nil {
		s.log.Error("failed to marshal notify payload", "error", err)
		return
	}

	req, err := http.NewRequest(http.MethodPost, s.notifyURL, bytes.NewBuffer(body))
	if err != nil {
		s.log.Error("failed to build notify request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.log.Error("failed to notify gateway", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		s.log.Warn("gateway notify returned non-2xx", "status", resp.StatusCode)
		return
	}

	s.log.Info("gateway notified about match", "user1", user1, "user2", user2, "chat_id", chatID, "match_hint", matchHint)
}
