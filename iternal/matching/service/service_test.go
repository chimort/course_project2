package matching

import (
	"context"
	"log/slog"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/chimort/course_project2/api/proto/matchingpb"
	"github.com/redis/go-redis/v9"
)

func setupService(t *testing.T) (*MatchingService, *redis.Client, context.Context) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis failed: %v", err)
	}
	t.Cleanup(func() { mr.Close() })

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	svc := NewMatchingService(rdb, nil, logg)
	return svc, rdb, context.Background()
}

// ----------------- 1 -----------------
func TestFindBestMatch_PicksBestByScore_Logged(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	users := []string{"alice", "bob", "carol"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{Username: u, Mode: matchingpb.MatchMode_MATCH_MODE_LANGUAGE})
		time.Sleep(1 * time.Millisecond)
	}

	profiles, _ := svc.fetchProfiles(ctx, users)
	me := profiles["alice"]

	type cs struct {
		id string
		s  float64
	}
	var list []cs
	for _, c := range []string{"bob", "carol"} {
		bd := CompatibilityDynamic(me, profiles[c], Preferences{})
		list = append(list, cs{id: c, s: bd.Total})
		t.Logf("alice -> %s: total=%.3f", c, bd.Total)
	}

	sort.Slice(list, func(i, j int) bool { return list[i].s > list[j].s })
	top := list[0].id

	chosen, err := svc.FindBestMatch(ctx, "alice", matchingpb.MatchMode_MATCH_MODE_LANGUAGE)
	if err != nil {
		t.Fatalf("FindBestMatch error: %v", err)
	}
	if chosen != top {
		t.Fatalf("expected best=%s, got %s", top, chosen)
	}
}

// ----------------- Complex profiles -----------------
func TestFindBestMatch_ComplexProfiles(t *testing.T) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	svc := NewMatchingService(rdb, nil, logg)

	ctx := context.Background()
	_ = rdb.FlushDB(ctx).Err()

	AllHobbies := []string{"reading", "movies", "gaming", "cooking", "yoga", "travel", "sports", "music", "coding", "painting"}
	meProfile := UserProfile{ID: "me", Age: 28, Hobbies: AllHobbies, Language: "English"}

	candidates := map[string]UserProfile{
		"a": {ID: "a", Age: 30, Hobbies: AllHobbies[:5], Language: "Spanish"},
		"b": {ID: "b", Age: 25, Hobbies: AllHobbies[5:], Language: "English"},
		"c": {ID: "c", Age: 28, Hobbies: []string{"coding", "reading"}, Language: "German"},
		"d": {ID: "d", Age: 35, Hobbies: []string{"travel", "music"}, Language: "French"},
	}

	usernames := []string{"me", "a", "b", "c", "d"}
	for _, u := range usernames {
		if _, err := svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{Username: u, Mode: matchingpb.MatchMode_MATCH_MODE_DEFAULT}); err != nil {
			t.Fatalf("JoinQueue %s failed: %v", u, err)
		}
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		out := map[string]UserProfile{"me": meProfile}
		for k, v := range candidates {
			out[k] = v
		}
		return out, nil
	})

	profiles, _ := svc.fetchProfiles(ctx, usernames)
	me := profiles["me"]

	for _, c := range []string{"a", "b", "c", "d"} {
		bd := CompatibilityDynamic(me, profiles[c], Preferences{WeightLanguage: 0.5, WeightHobbies: 0.4, WeightAge: 0.1})
		t.Logf("me -> %s: total=%.3f (lang=%.3f hobby=%.3f age=%.3f)", c, bd.Total, bd.WeightedLang, bd.WeightedHobby, bd.WeightedAge)
	}

	chosen, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_DEFAULT)
	if err != nil {
		t.Fatalf("FindBestMatch error: %v", err)
	}
	if chosen == "" {
		t.Fatalf("expected a match, got empty")
	}
	t.Logf("FindBestMatch chose: %s", chosen)
}