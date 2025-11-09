package handlers

import (
	"context"
	"encoding/json"
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
	h.log.Info("profile successfully fetched", "username", username)
	c.Response().Header().Set("Content-Type", "application/json")
	return json.NewEncoder(c.Response()).Encode(res.User)
}