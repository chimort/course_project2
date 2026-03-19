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

func lang(name, level string) []LanguageSkill {
	return []LanguageSkill{{Name: name, Level: level}}
}

func setupService(t *testing.T) (*MatchingService, *redis.Client, context.Context) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis failed: %v", err)
	}
	t.Cleanup(func() { mr.Close() })

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	svc := NewMatchingService(rdb, nil, nil, logg)
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
			Username:     u,
			Mode:         matchingpb.MatchMode_MATCH_MODE_LANGUAGE,
			LanguageMode: matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_LEARNING_GOALS,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"alice": {ID: "alice", Age: 28, Hobbies: []string{"movies", "reading"}, Languages: append(lang("English", "NATIVE"), LanguageSkill{Name: "Russian", Level: "LOW"})},
			"bob":   {ID: "bob", Age: 35, Hobbies: []string{"cooking"}, Languages: append(lang("Spanish", "NATIVE"), LanguageSkill{Name: "English", Level: "LOW"})},
			"carol": {ID: "carol", Age: 27, Hobbies: []string{"movies", "reading"}, Languages: append(lang("English", "LOW"), LanguageSkill{Name: "Russian", Level: "NATIVE"})},
		}, nil
	})

	res, err := svc.findBestMatchWithSeen(
		ctx,
		"alice",
		matchingpb.MatchMode_MATCH_MODE_LANGUAGE,
		matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_LEARNING_GOALS,
		nil,
	)
	if err != nil {
		t.Fatalf("FindBestMatch error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if res.Username != "carol" {
		t.Fatalf("expected best=carol for language exchange, got %s", res.Username)
	}
}

func TestFindBestMatch_LanguageMode_SameLanguage(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	users := []string{"alice", "bob", "carol"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username:     u,
			Mode:         matchingpb.MatchMode_MATCH_MODE_LANGUAGE,
			LanguageMode: matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"alice": {ID: "alice", Age: 28, Hobbies: []string{"music"}, Languages: []LanguageSkill{{Name: "English", Level: "NATIVE"}, {Name: "Russian", Level: "LOW"}}},
			"bob":   {ID: "bob", Age: 29, Hobbies: []string{"books"}, Languages: []LanguageSkill{{Name: "English", Level: "MEDIUM"}}},
			"carol": {ID: "carol", Age: 27, Hobbies: []string{"movies"}, Languages: []LanguageSkill{{Name: "Russian", Level: "NATIVE"}}},
		}, nil
	})

	res, err := svc.findBestMatchWithSeen(ctx, "alice", matchingpb.MatchMode_MATCH_MODE_LANGUAGE, matchingpb.LanguageMatchMode_LANGUAGE_MATCH_MODE_SAME_LANGUAGE, nil)
	if err != nil {
		t.Fatalf("FindBestMatch error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected non-nil result")
	}
	if res.Username != "bob" {
		t.Fatalf("expected best=bob for same-language search, got %s", res.Username)
	}
}

//
// ----------------- 2. DISCUSS MOVIE -----------------
//

func TestFindBestMatch_DiscussMovie_Basic(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me":    {ID: "me", Age: 28, Hobbies: []string{"movies", "reading"}, Languages: lang("English", "NATIVE")},
			"m1":    {ID: "m1", Age: 26, Hobbies: []string{"movies", "gaming"}, Languages: lang("English", "NATIVE")},
			"other": {ID: "other", Age: 30, Hobbies: []string{"cooking"}, Languages: lang("English", "NATIVE")},
		}, nil
	})

	users := []string{"me", "m1", "other"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
		})
		time.Sleep(1 * time.Millisecond)
	}

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

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me": {ID: "me", Age: 28, Hobbies: []string{"movies"}, Languages: lang("English", "NATIVE")},
			"u1": {ID: "u1", Age: 28, Hobbies: []string{"cooking"}, Languages: lang("English", "NATIVE")},
		}, nil
	})

	users := []string{"me", "u1"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
		})
		time.Sleep(1 * time.Millisecond)
	}

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

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me":       {ID: "me", Age: 28, Hobbies: []string{"movies"}, Languages: lang("English", "NATIVE")},
			"sameLang": {ID: "sameLang", Age: 35, Hobbies: []string{"movies"}, Languages: lang("English", "NATIVE")},
			"diffLang": {ID: "diffLang", Age: 28, Hobbies: []string{"movies"}, Languages: lang("Spanish", "NATIVE")},
		}, nil
	})

	users := []string{"me", "sameLang", "diffLang"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
		})
		time.Sleep(1 * time.Millisecond)
	}

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
			"me": {ID: "me", Age: 28, Hobbies: []string{"movies", "travel"}, Languages: lang("English", "NATIVE")},
			"a":  {ID: "a", Age: 35, Hobbies: []string{"cooking"}, Languages: lang("Spanish", "NATIVE")},
			"b":  {ID: "b", Age: 27, Hobbies: []string{"travel"}, Languages: lang("English", "NATIVE")},
			"c":  {ID: "c", Age: 50, Hobbies: []string{"gaming"}, Languages: lang("German", "NATIVE")},
		}, nil
	})

	res, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_DEFAULT)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res == nil {
		t.Fatalf("expected match")
	}
	if res.Reason != "best match: travel" {
		t.Fatalf("expected travel to be preferred over language, got %q", res.Reason)
	}
}

func TestFindBestMatch_DiscussMusic_Basic(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me":       {ID: "me", Age: 24, Hobbies: []string{"music", "books"}, Languages: lang("English", "NATIVE")},
			"musicFan": {ID: "musicFan", Age: 25, Hobbies: []string{"music", "sport"}, Languages: lang("English", "NATIVE")},
			"reader":   {ID: "reader", Age: 24, Hobbies: []string{"books"}, Languages: lang("English", "NATIVE")},
		}, nil
	})

	users := []string{"me", "musicFan", "reader"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MUSIC,
		})
		time.Sleep(1 * time.Millisecond)
	}

	res, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_DISCUSS_MUSIC)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res == nil || res.Username != "musicFan" {
		t.Fatalf("expected musicFan, got %+v", res)
	}
	if res.Reason != "best match: music" {
		t.Fatalf("expected 'best match: music', got %q", res.Reason)
	}
}

func TestFindBestMatch_Fast_Basic(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	users := []string{"me", "near", "far"}
	for _, u := range users {
		_, _ = svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
			Username: u,
			Mode:     matchingpb.MatchMode_MATCH_MODE_FAST,
		})
		time.Sleep(1 * time.Millisecond)
	}

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me":   {ID: "me", Age: 24, Hobbies: []string{"books"}, Languages: []LanguageSkill{{Name: "English", Level: "NATIVE"}}},
			"near": {ID: "near", Age: 27, Hobbies: []string{"sport"}, Languages: []LanguageSkill{{Name: "English", Level: "MEDIUM"}}},
			"far":  {ID: "far", Age: 41, Hobbies: []string{"music"}, Languages: []LanguageSkill{{Name: "English", Level: "MEDIUM"}}},
		}, nil
	})

	res, err := svc.FindBestMatch(ctx, "me", matchingpb.MatchMode_MATCH_MODE_FAST)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if res == nil || res.Username != "near" {
		t.Fatalf("expected near, got %+v", res)
	}
	if res.Reason != "best match: fast_chat" {
		t.Fatalf("expected fast_chat reason, got %q", res.Reason)
	}
}

func TestJoinQueue_RejectsUnavailableInterestMode(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me": {ID: "me", Age: 24, Hobbies: []string{"books"}, Languages: lang("English", "NATIVE")},
		}, nil
	})

	res, err := svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
		Username: "me",
		Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Ok {
		t.Fatalf("expected join to be rejected for missing interest")
	}

	users, err := svc.ListQueue(ctx, &matchingpb.ListQueueRequest{})
	if err != nil {
		t.Fatalf("list queue failed: %v", err)
	}
	if len(users.Usernames) != 0 {
		t.Fatalf("expected queue to stay empty, got %+v", users.Usernames)
	}
}

func TestJoinQueue_AllowsOwnedInterestMode(t *testing.T) {
	svc, rdb, ctx := setupService(t)
	_ = rdb.FlushDB(ctx).Err()

	svc.SetFetchFunc(func(ctx context.Context, usernames []string) (map[string]UserProfile, error) {
		return map[string]UserProfile{
			"me": {ID: "me", Age: 24, Hobbies: []string{"movies", "books"}, Languages: lang("English", "NATIVE")},
		}, nil
	})

	res, err := svc.JoinQueue(ctx, &matchingpb.JoinQueueRequest{
		Username: "me",
		Mode:     matchingpb.MatchMode_MATCH_MODE_DISCUSS_MOVIE,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Ok {
		t.Fatalf("expected join to succeed for owned interest")
	}
}
