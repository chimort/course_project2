package token

import (
	"errors"
	"time"

	"github.com/chimort/course_project2/iternal/auth/utils"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrExpired     = errors.New("access token expired")
	ErrInvalid     = errors.New("invalid token")
	ErrRefreshOnly = errors.New("access token expired, refresh valid")
)

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func generateToken(username string, secret []byte, duration time.Duration) (string, error) {
	expTime := time.Now().Add(duration)
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func GenerateJwt(username string) (string, error) {
	return generateToken(username, utils.JWTKEY, 20*time.Minute)
}

func GenerateRefreshToken(username string) (string, error) {
	return generateToken(username, utils.REFRESHKEY, 7*24*time.Hour)
}

func validateToken(tokenStr string, secret []byte) (*Claims, error) {
	if tokenStr == "" {
		return nil, ErrInvalid
	}

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
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
	return validateToken(tokenStr, utils.REFRESHKEY)
}

func ValidateToken(tokenStr string) (*Claims, error) {
	return validateToken(tokenStr, utils.JWTKEY)
}
