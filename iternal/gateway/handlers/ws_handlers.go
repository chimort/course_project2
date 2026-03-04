package handlers

import (
	"log/slog"
	"net/http"

	gw "github.com/chimort/course_project2/iternal/gateway/websocket"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WSHandler struct {
	hub      *gw.WSHub
	log      *slog.Logger
	upgrader websocket.Upgrader
}

type NotifyMatchRequest struct {
	User1  string `json:"user1"`
	User2  string `json:"user2"`
	ChatID string `json:"chat_id"`
}

func NewWSHandler(hub *gw.WSHub, log *slog.Logger) *WSHandler {
	return &WSHandler{
		hub: hub,
		log: log.With("component", "ws_handler"),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (h *WSHandler) HandleWS(c echo.Context) error {
	username := c.QueryParam("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username required"})
	}

	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		h.log.Error("failed to upgrade ws", "error", err)
		return err
	}

	h.hub.Set(username, conn)
	h.log.Info("ws client connected", "username", username)

	defer func() {
		h.hub.Remove(username)
		h.log.Info("ws client disconnected", "username", username)
	}()

	// Пока просто держим соединение живым.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	return nil
}

func (h *WSHandler) NotifyMatchFound(c echo.Context) error {
	var req NotifyMatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad payload"})
	}

	if req.User1 == "" || req.User2 == "" || req.ChatID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
	}

	_ = h.hub.Send(req.User1, map[string]string{
		"type":    "match_found",
		"chat_id": req.ChatID,
		"partner": req.User2,
	})

	_ = h.hub.Send(req.User2, map[string]string{
		"type":    "match_found",
		"chat_id": req.ChatID,
		"partner": req.User1,
	})

	h.log.Info("match event sent", "user1", req.User1, "user2", req.User2, "chat_id", req.ChatID)
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}