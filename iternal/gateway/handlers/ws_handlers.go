package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	fastChats  map[string]fastChatSession
	fastMu     sync.RWMutex
}

type NotifyMatchRequest struct {
	User1     string `json:"user1"`
	User2     string `json:"user2"`
	ChatID    string `json:"chat_id"`
	MatchHint string `json:"match_hint"`
	FastChat  bool   `json:"fast_chat"`
}

type ChatMessage struct {
	Type            string `json:"type"`
	ChatID          string `json:"chat_id"`
	Text            string `json:"text"`
	MessageType     string `json:"message_type,omitempty"`
	FileURL         string `json:"file_url,omitempty"`
	FileName        string `json:"file_name,omitempty"`
	MimeType        string `json:"mime_type,omitempty"`
	FileSizeBytes   int64  `json:"file_size_bytes,omitempty"`
	DurationSeconds int32  `json:"duration_seconds,omitempty"`
}

type fastChatSession struct {
	User1     string
	User2     string
	CreatedAt time.Time
}

func NewWSHandler(hub *gw.WSHub, log *slog.Logger, chatClient chatpb.ChatServiceClient) *WSHandler {
	return &WSHandler{
		hub: hub,
		log: log.With("component", "ws_handler"),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		chatClient: chatClient,
		fastChats:  make(map[string]fastChatSession),
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
		h.hub.Remove(username, conn)
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
	if participants, ok := h.getFastChatParticipants(msg.ChatID); ok {
		for _, participant := range participants {
			if participant != sender {
				_ = h.hub.Send(participant, map[string]interface{}{
					"type":      "chat_message",
					"from":      sender,
					"text":      msg.Text,
					"chat_id":   msg.ChatID,
					"fast_chat": true,
				})
			}
		}

		h.log.Info("fast chat message relayed", "sender", sender, "chat_id", msg.ChatID)
		return
	}

	// Send message via gRPC
	_, err := h.chatClient.SendMessage(context.Background(), &chatpb.SendMessageRequest{
		ChatId:          msg.ChatID,
		Sender:          sender,
		Content:         msg.Text,
		MessageType:     chooseMessageType(msg.MessageType),
		FileUrl:         msg.FileURL,
		FileName:        msg.FileName,
		MimeType:        msg.MimeType,
		FileSizeBytes:   msg.FileSizeBytes,
		DurationSeconds: msg.DurationSeconds,
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
				"type":             "chat_message",
				"from":             sender,
				"text":             msg.Text,
				"chat_id":          msg.ChatID,
				"message_type":     chooseMessageType(msg.MessageType),
				"file_url":         msg.FileURL,
				"file_name":        msg.FileName,
				"mime_type":        msg.MimeType,
				"file_size_bytes":  msg.FileSizeBytes,
				"duration_seconds": msg.DurationSeconds,
			})
		}
	}

	h.log.Info("chat message sent", "sender", sender, "chat_id", msg.ChatID)
}

func chooseMessageType(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "image", "file", "voice":
		return strings.ToLower(strings.TrimSpace(kind))
	default:
		return "text"
	}
}

func (h *WSHandler) UploadAttachment(c echo.Context) error {
	chatID := strings.TrimSpace(c.FormValue("chat_id"))
	sender := strings.TrimSpace(c.FormValue("sender"))
	durationRaw := strings.TrimSpace(c.FormValue("duration_seconds"))
	if chatID == "" || sender == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "chat_id and sender required"})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "file required"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "cannot open file"})
	}
	defer src.Close()

	if err := os.MkdirAll("web/static/uploads/chat", 0o755); err != nil {
		h.log.Error("failed to create uploads dir", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot prepare upload dir"})
	}

	safeName := sanitizeFileName(fileHeader.Filename)
	storedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	relativePath := filepath.ToSlash(filepath.Join("uploads", "chat", storedName))
	dstPath := filepath.Join("web", "static", "uploads", "chat", storedName)

	dst, err := os.Create(dstPath)
	if err != nil {
		h.log.Error("failed to create uploaded file", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot save file"})
	}
	defer dst.Close()

	size, err := copyUploadedFile(dst, src)
	if err != nil {
		h.log.Error("failed to write uploaded file", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot save file"})
	}

	messageType := detectMessageType(c.FormValue("message_type"), fileHeader.Header.Get("Content-Type"), safeName)
	duration := parseDurationSeconds(durationRaw)
	fileURL := "/static/" + relativePath

	_, err = h.chatClient.SendMessage(context.Background(), &chatpb.SendMessageRequest{
		ChatId:          chatID,
		Sender:          sender,
		Content:         "",
		MessageType:     messageType,
		FileUrl:         fileURL,
		FileName:        safeName,
		MimeType:        fileHeader.Header.Get("Content-Type"),
		FileSizeBytes:   size,
		DurationSeconds: duration,
	})
	if err != nil {
		h.log.Error("failed to persist attachment message", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "cannot save attachment message"})
	}

	participants, err := h.chatClient.GetParticipants(context.Background(), &chatpb.GetParticipantsRequest{ChatId: chatID})
	if err == nil {
		for _, participant := range participants.Usernames {
			if participant == sender {
				continue
			}
			_ = h.hub.Send(participant, map[string]interface{}{
				"type":             "chat_message",
				"from":             sender,
				"text":             "",
				"chat_id":          chatID,
				"message_type":     messageType,
				"file_url":         fileURL,
				"file_name":        safeName,
				"mime_type":        fileHeader.Header.Get("Content-Type"),
				"file_size_bytes":  size,
				"duration_seconds": duration,
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"ok":               true,
		"message_type":     messageType,
		"file_url":         fileURL,
		"file_name":        safeName,
		"mime_type":        fileHeader.Header.Get("Content-Type"),
		"file_size_bytes":  size,
		"duration_seconds": duration,
	})
}

func sanitizeFileName(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	base = strings.ReplaceAll(base, " ", "_")
	base = strings.ReplaceAll(base, "..", "")
	if base == "" {
		return "upload.bin"
	}
	return base
}

func copyUploadedFile(dst *os.File, src io.Reader) (int64, error) {
	return io.Copy(dst, src)
}

func detectMessageType(requested, mimeType, fileName string) string {
	req := strings.ToLower(strings.TrimSpace(requested))
	mime := strings.ToLower(strings.TrimSpace(mimeType))
	if req == "voice" || strings.HasPrefix(mime, "audio/") {
		return "voice"
	}
	if req == "image" || strings.HasPrefix(mime, "image/") {
		return "image"
	}
	if strings.HasSuffix(strings.ToLower(fileName), ".png") || strings.HasSuffix(strings.ToLower(fileName), ".jpg") || strings.HasSuffix(strings.ToLower(fileName), ".jpeg") || strings.HasSuffix(strings.ToLower(fileName), ".gif") || strings.HasSuffix(strings.ToLower(fileName), ".webp") {
		return "image"
	}
	return "file"
}

func parseDurationSeconds(raw string) int32 {
	if raw == "" {
		return 0
	}
	var seconds int32
	fmt.Sscanf(raw, "%d", &seconds)
	return seconds
}

func (h *WSHandler) NotifyMatchFound(c echo.Context) error {
	var req NotifyMatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "bad payload"})
	}

	if req.User1 == "" || req.User2 == "" || req.ChatID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing fields"})
	}

	if req.FastChat {
		h.registerFastChat(req.ChatID, req.User1, req.User2)
	}

	_ = h.hub.Send(req.User1, map[string]interface{}{
		"type":       "match_found",
		"chat_id":    req.ChatID,
		"partner":    req.User2,
		"match_hint": req.MatchHint,
		"fast_chat":  req.FastChat,
	})

	_ = h.hub.Send(req.User2, map[string]interface{}{
		"type":       "match_found",
		"chat_id":    req.ChatID,
		"partner":    req.User1,
		"match_hint": req.MatchHint,
		"fast_chat":  req.FastChat,
	})

	h.log.Info("match event sent", "user1", req.User1, "user2", req.User2, "chat_id", req.ChatID, "fast_chat", req.FastChat)
	return c.JSON(http.StatusOK, map[string]bool{"ok": true})
}

func (h *WSHandler) registerFastChat(chatID, user1, user2 string) {
	h.fastMu.Lock()
	defer h.fastMu.Unlock()

	h.fastChats[chatID] = fastChatSession{
		User1:     user1,
		User2:     user2,
		CreatedAt: time.Now(),
	}
}

func (h *WSHandler) getFastChatParticipants(chatID string) ([]string, bool) {
	h.fastMu.RLock()
	session, ok := h.fastChats[chatID]
	h.fastMu.RUnlock()
	if !ok {
		return nil, false
	}

	if time.Since(session.CreatedAt) > 12*time.Hour {
		h.fastMu.Lock()
		delete(h.fastChats, chatID)
		h.fastMu.Unlock()
		return nil, false
	}

	return []string{session.User1, session.User2}, true
}
