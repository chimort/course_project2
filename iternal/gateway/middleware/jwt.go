package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/chimort/course_project2/iternal/auth/token"
	"github.com/labstack/echo/v4"
)

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing token"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token format"})
		}

		accessToken := parts[1]
		refreshToken := c.Request().Header.Get("X-Refresh-Token")

		claims, err := token.ValidateToken(accessToken)
		if err == nil {
			c.Set("username", claims.Username)
			return next(c)
		}

		if errors.Is(err, token.ErrExpired) {
			refreshClaims, rErr := token.ValidateToken(refreshToken)
			if rErr == nil {
				c.Set("username", refreshClaims.Username)
				return next(c)
			}

			if errors.Is(rErr, token.ErrExpired) {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "both tokens expired"})
			}
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid refresh token"})
		}

		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid access token"})
	}
}

