package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("secretkey")
var refreshKey = []byte("secretrefreshkey")

var (
	ErrExpired     = errors.New("access token expired")
	ErrInvalid     = errors.New("invalid token")
	ErrRefreshOnly = errors.New("access token expired, refresh valid")
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateJwt(username string) (string, error) {
	expTime := time.Now().Add(12 * time.Hour)

	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func GenerateRefreshToken(username string) (string, error) {
	expiration := time.Now().Add(7 * 24 * time.Hour)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshKey)
}

func ValidateAccessToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, ErrInvalid
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalid
}

func ValidateRefreshToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return refreshKey, nil
	})
	if err != nil {
		return nil, ErrInvalid
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalid
}

func ValidateTokens(accessToken, refreshToken string) (*Claims, error) {
	claims, err := ValidateAccessToken(accessToken)
	if err == nil {
		return claims, nil
	}

	if errors.Is(err, ErrExpired) && refreshToken != "" {
		refreshClaims, rErr := ValidateRefreshToken(refreshToken)
		if rErr == nil {
			return refreshClaims, ErrRefreshOnly
		}
		return nil, ErrInvalid
	}

	return nil, err
}
