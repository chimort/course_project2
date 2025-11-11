package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/chimort/course_project2/iternal/auth/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const UsernameKey ctxKey = "username"

func AuthUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		if val := md.Get("internal"); len(val) > 0 && val[0] == "true" {
    		return handler(ctx, req)
		}

		authHeader := md["authorization"]
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		parts := strings.Split(authHeader[0], " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return nil, status.Error(codes.Unauthenticated, "invalid token format")
		}

		accessToken := parts[1]

		claims, err := token.ValidateToken(accessToken)
		if err != nil {
			refreshHeader := md["x-refresh-token"]
			if len(refreshHeader) > 0 {
				refreshClaims, rErr := token.ValidateToken(refreshHeader[0])
				if rErr == nil {
					newCtx := context.WithValue(ctx, UsernameKey, refreshClaims.Username)
					return handler(newCtx, req)
				}
				if errors.Is(rErr, token.ErrExpired) {
					return nil, status.Error(codes.Unauthenticated, "both tokens expired")
				}
				return nil, status.Error(codes.Unauthenticated, "invalid refresh token")
			}
			if errors.Is(err, token.ErrExpired) {
				return nil, status.Error(codes.Unauthenticated, "access token expired")
			}
			return nil, status.Error(codes.Unauthenticated, "invalid access token")
		}

		newCtx := context.WithValue(ctx, UsernameKey, claims.Username)
		return handler(newCtx, req)
	}
}