package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/chimort/course_project2/iternal/auth/token"
	"github.com/chimort/course_project2/iternal/auth/utils"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func expiredAccessToken(t *testing.T, username string) string {
	t.Helper()

	claims := &token.Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(utils.JWTKEY)
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}
	return signed
}

func TestAuthUnaryInterceptor_AllowsValidRefreshWhenAccessExpired(t *testing.T) {
	access := expiredAccessToken(t, "alice")
	refresh, err := token.GenerateRefreshToken("alice")
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+access,
		"x-refresh-token", refresh,
	))

	interceptor := AuthUnaryInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	resp, err := interceptor(ctx, "req", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		username, _ := ctx.Value(UsernameKey).(string)
		return username, nil
	})
	if err != nil {
		t.Fatalf("expected request to pass with valid refresh token, got error: %v", err)
	}

	if resp != "alice" {
		t.Fatalf("expected username from refresh token context, got %v", resp)
	}
}

func TestAuthUnaryInterceptor_RejectsInvalidRefreshToken(t *testing.T) {
	access := expiredAccessToken(t, "alice")
	wrongRefresh, err := token.GenerateJwt("alice")
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(
		"authorization", "Bearer "+access,
		"x-refresh-token", wrongRefresh,
	))

	interceptor := AuthUnaryInterceptor()
	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/Method"}

	_, err = interceptor(ctx, "req", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "unexpected", nil
	})
	if err == nil {
		t.Fatalf("expected invalid refresh token error")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated, got %v", status.Code(err))
	}
}
