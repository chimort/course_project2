package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/chimort/course_project2/api/proto/chatpb"
	gw "github.com/chimort/course_project2/iternal/gateway/websocket"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WSHandler struct {
	hub        *gw.WSHub
	log        *slog.Logger
	upgrader   websocket.Upgrader
	chatClient chatpb.ChatServiceClient
}

type NotifyMatchRequest struct {
	User1  string `json:"user1"`
	User2  string `json:"user2"`
	ChatID string `json:"chat_id"`
}

type ChatMessage struct {
	Type   string `json:"type"`
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func NewWSHandler(hub *gw.WSHub, log *slog.Logger, chatClient chatpb.ChatServiceClient) *WSHandler {
	return &WSHandler{
		hub: hub,
		log: log.With("component", "ws_handler"),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		chatClient: chatClient,
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

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg ChatMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			h.log.Error("failed to parse ws message", "error", err, "data", string(data))
			continue
		}

		if msg.Type == "chat_message" {
			h.handleChatMessage(username, &msg)
		}
	}

	return nil
}

func (h *WSHandler) handleChatMessage(sender string, msg *ChatMessage) {
	// Send message via gRPC
	_, err := h.chatClient.SendMessage(context.Background(), &chatpb.SendMessageRequest{
		ChatId:  msg.ChatID,
		Sender:  sender,
		Content: msg.Text,
	})
	if err != nil {
		h.log.Error("failed to send message", "error", err)
		return
	}

	// Get participants
	participants, err := h.chatClient.GetParticipants(context.Background(), &chatpb.GetParticipantsRequest{
		ChatId: msg.ChatID,
	})
	if err != nil {
		h.log.Error("failed to get participants", "error", err)
		return
	}

	// Send to other participants
	for _, participant := range participants.Usernames {
		if participant != sender {
			_ = h.hub.Send(participant, map[string]interface{}{
				"type":    "chat_message",
				"from":    sender,
				"text":    msg.Text,
				"chat_id": msg.ChatID,
			})
		}
	}

	h.log.Info("chat message sent", "sender", sender, "chat_id", msg.ChatID)
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
