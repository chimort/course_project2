package middleware

import (
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

		claims, err := token.ValidateAccessToken(parts[1])
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "expired token"})
		}

		c.Set("user_id", claims.UserId)
		c.Set("username", claims.Username)
		return next(c)
	}
}
