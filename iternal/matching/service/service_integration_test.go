package matching

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/chimort/course_project2/api/proto/matchingpb"
	"github.com/redis/go-redis/v9"
)

// Интеграционный тест для MatchingService с in-memory redis (miniredis).
func TestMatchingService_JoinFindLeaveFlow(t *testing.T) {
	// поднять in-memory redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	svc := NewMatchingService(rdb, nil, nil, logg)
	ctx := context.Background()
	_ = rdb.FlushDB(ctx).Err()

	// Добавим троих пользователей в очередь
	users := []string{"alice", "bob", "carol"}
	for _, u := range users {
		_, err := svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username:     u,
			Mode:         matchingpb.MatchMode_MATCH_MODE_LANGUAGE,
			LanguageMode: matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_LEARNING_GOALS,
		})
		if err != nil {
			t.Fatalf("JoinQueue(%s) failed: %v", u, err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Проверим ListQueue
	listResp, err := svc.ListQueue(ctx, &matchingpb.ListQueueRequest{})
	if err != nil {
		t.Fatalf("ListQueue failed: %v", err)
	}
	if len(listResp.Usernames) != len(users) {
		t.Fatalf("expected %d users in queue, got %d (%v)", len(users), len(listResp.Usernames), listResp.Usernames)
	}

	// Подменяем fetchProfiles так, чтобы была понятная логика (hardcode)
	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		out := map[string]UserProfile{
			"alice": {ID: "alice", Age: 28, Hobbies: []string{"movies"}, Languages: []LanguageSkill{{Name: "English", Level: "NATIVE"}, {Name: "Russian", Level: "LOW"}}},
			"bob":   {ID: "bob", Age: 30, Hobbies: []string{"movies", "reading"}, Languages: []LanguageSkill{{Name: "English", Level: "LOW"}, {Name: "Russian", Level: "NATIVE"}}},
			"carol": {ID: "carol", Age: 40, Hobbies: []string{"cooking"}, Languages: []LanguageSkill{{Name: "Spanish", Level: "NATIVE"}}},
		}
		return out, nil
	})

	// Найдём матч для alice
	res, err := svc.findBestMatchWithSeen(
		ctx,
		"alice",
		matchingpb.MatchMode_MATCH_MODE_LANGUAGE,
		matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_LEARNING_GOALS,
		nil,
	)
	if err != nil {
		t.Fatalf("FindBestMatch failed: %v", err)
	}
	if res == nil {
		t.Fatalf("expected a match for alice, got empty")
	}
	if res.Username == "alice" {
		t.Fatalf("match for alice is alice (self), unexpected")
	}

	// После матчмейкинга alice и найденный соперник должны быть удалены из очереди
	listResp2, err := svc.ListQueue(ctx, &matchingpb.ListQueueRequest{})
	if err != nil {
		t.Fatalf("ListQueue after match failed: %v", err)
	}
	if len(listResp2.Usernames) >= len(listResp.Usernames) {
		t.Fatalf("expected queue to shrink after matching; before=%d after=%d", len(listResp.Usernames), len(listResp2.Usernames))
	}

	// Тест LeaveQueue: положим одного обратно и удалим
	_, err = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{Username: "dave", Mode: matchingpb.MatchMode_MATCH_MODE_DEFAULT})
	if err != nil {
		t.Fatalf("JoinQueue(dave) failed: %v", err)
	}
	_, err = svc.LeaveQueue(ctx, &matchingpb.LeaveQueueRequest{Username: "dave"})
	if err != nil {
		t.Fatalf("LeaveQueue(dave) failed: %v", err)
	}
}
