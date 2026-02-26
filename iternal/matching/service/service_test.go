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

//
// ----------------- 1. LANGUAGE MODE -----------------
//

func TestFindBestMatch_PicksBestByScore_Logged(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	users := []string{"alice", "bob", "carol"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_LANGUAGE,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"alice": {ID: "alice", Age: 28, Hobbies: []string{"movies", "reading"}, Language: "English"},
			"bob":   {ID: "bob", Age: 35, Hobbies: []string{"cooking"}, Language: "Spanish"},
			"carol": {ID: "carol", Age: 27, Hobbies: []string{"movies", "reading"}, Language: "English"},
		}, nil
	})

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
	}

	sort.Slice(list, func(i, j int) bool { return list[i].s > list[j].s })
	expectedBest := list[0].id

	res, err := svc.FindBestMatch(ctx, "alice", matchingpb.MatchMode_MATCH_MODE_LANGUAGE)
	if err != nil {
		t.Fatalf("FindBestMatch error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if res.Username != expectedBest {
		t.Fatalf("expected best=%s, got %s", expectedBest, res.Username)
	}
}

//
// ----------------- 2. DISCUSS MOVIE -----------------
//

func TestFindBestMatch_DiscussMovie_Basic(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	users := []string{"me", "m1", "other"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me":    {ID: "me", Age: 28, Hobbies: []string{"movies", "reading"}, Language: "English"},
			"m1":    {ID: "m1", Age: 26, Hobbies: []string{"movies", "gaming"}, Language: "English"},
			"other": {ID: "other", Age: 30, Hobbies: []string{"cooking"}, Language: "English"},
		}, nil
	})

	res, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE)
	if err != nil {
		t.Fatalf("FindBestMatch error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected match result, got nil")
	}
	if res.Username != "m1" {
		t.Fatalf("expected m1, got %s", res.Username)
	}
	if res.Reason != "best match: movies" {
		t.Fatalf("expected 'best match: movies', got %q", res.Reason)
	}
}

//
// ---------- NEW: No one with movies ----------
//

func TestFindBestMatch_DiscussMovie_NoCandidates(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	users := []string{"me", "u1"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me": {ID: "me", Age: 28, Hobbies: []string{"movies"}, Language: "English"},
			"u1": {ID: "u1", Age: 28, Hobbies: []string{"cooking"}, Language: "English"},
		}, nil
	})

	res, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res != nil {
		t.Fatalf("expected nil match, got %+v", res)
	}
}

//
// ---------- NEW: Language priority inside movie mode ----------
//

func TestFindBestMatch_DiscussMovie_LanguagePriority(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	users := []string{"me", "sameLang", "diffLang"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me":       {ID: "me", Age: 28, Hobbies: []string{"movies"}, Language: "English"},
			"sameLang": {ID: "sameLang", Age: 35, Hobbies: []string{"movies"}, Language: "English"},
			"diffLang": {ID: "diffLang", Age: 28, Hobbies: []string{"movies"}, Language: "Spanish"},
		}, nil
	})

	res, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected match")
	}
	if res.Username != "sameLang" {
		t.Fatalf("expected sameLang due to language priority, got %s", res.Username)
	}
}

//
// ----------------- 3. DEFAULT COMPLEX -----------------
//

func TestFindBestMatch_DefaultComplex(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	all := []string{"me", "a", "b", "c"}
	for _, u := range all {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DEFAULT,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me": {ID: "me", Age: 28, Hobbies: []string{"movies", "travel"}, Language: "English"},
			"a":  {ID: "a", Age: 35, Hobbies: []string{"cooking"}, Language: "Spanish"},
			"b":  {ID: "b", Age: 27, Hobbies: []string{"travel"}, Language: "English"},
			"c":  {ID: "c", Age: 50, Hobbies: []string{"gaming"}, Language: "German"},
		}, nil
	})

	res, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_DEFAULT)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected match")
	}
}