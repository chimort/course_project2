package matching

import (
	"context"
	"os"
	"testing"
	"time"
	"log/slog"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/chimort/course_project2/api/proto/matchingpb"
)

// Интеграционный тест для MatchingService с in-memory redis (miniredis).
func TestMatchingService_JoinFindLeaveFlow(t *testing.T) {
	// поднять in-memory redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	defer mr.Close()

	// подключение go-redis к miniredis
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// простой logger
	logg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	// создаём сервис (userClient = nil => он будет использовать встроенный хардкод в fetchProfiles)
	svc := NewMatchingService(rdb, nil, logg)

	ctx := context.Background()

	// Очистим (на всякий)
	_ = rdb.FlushDB(ctx).Err()

	// Добавим троих пользователей в очередь
	users := []string{"alice", "bob", "carol"}
	for _, u := range users {
		_, err := svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_LANGUAGE,
		})
		if err != nil {
			t.Fatalf("JoinQueue(%s) failed: %v", u, err)
		}
		// небольшая пауза чтобы score различался
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

	// Попробуем найти для alice лучший матч (в хардкоде fetchProfiles у нас bob лучше)
	match, err := svc.FindBestMatch(ctx, "alice", matchingpb.MatchMode_MATCH_MODE_LANGUAGE)
	if err != nil {
		t.Fatalf("FindBestMatch failed: %v", err)
	}
	if match == "" {
		t.Fatalf("expected a match for alice, got empty")
	}
	// В хардкоде мы возвращаем bob/others — просто проверим, что вернулся не сам alice
	if match == "alice" {
		t.Fatalf("match for alice is alice (self), unexpected")
	}

	// После матчмейкинга alice и найденный соперник должны быть удалены из очереди
	listResp2, err := svc.ListQueue(ctx, &matchingpb.ListQueueRequest{})
	if err != nil {
		t.Fatalf("ListQueue after match failed: %v", err)
	}
	// ожидание: осталось как минимум 1 пользователь
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