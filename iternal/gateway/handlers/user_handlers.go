package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"


	"github.com/chimort/course_project2/api/proto/userpb"
	"github.com/labstack/echo/v4"
)


type UserHandler struct {
	UserClient userpb.UserServiceClient
	log  *slog.Logger
}

func NewUserHandler(client userpb.UserServiceClient, log *slog.Logger) *UserHandler {
	return &UserHandler{
		UserClient: client,
		log: log,
	}
}

func (h *UserHandler) GetProfile(c echo.Context) error {
	username, ok := c.Get("username").(string)
	if !ok || username == "" {
		h.log.Warn("missing username in context")
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	h.log.Info("profile request recived", "username", username)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := h.UserClient.GetUser(ctx, &userpb.GetUserRequest{Username: username,})
	if err != nil {
		h.log.Error("failed to get user", "username", username, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	profile := UserProfileResponse{
		Username:  res.User.Username,
        Email:     res.User.Email,
        Age:       res.User.Age,
        Gender:    res.User.Gender,
        Languages: res.User.Languages,
		Interests: res.User.Interests,
	}

	h.log.Info("profile successfully fetched", "username", username)
	return c.JSON(http.StatusOK, profile)
}
