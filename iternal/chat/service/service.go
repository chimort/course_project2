package chat

import (
	"context"
	"log/slog"

	"github.com/chimort/course_project2/iternal/chat/models"
	"github.com/chimort/course_project2/iternal/chat/repository"
)

type ChatService struct {
	repo *repository.ChatRepository
	log  *slog.Logger
}

func NewChatService(repo *repository.ChatRepository, log *slog.Logger) *ChatService {
	return &ChatService{
		repo: repo,
		log:  log.With("component", "chat_service"),
	}
}

func (s *ChatService) CreateChat(ctx context.Context, user1, user2 string) (*models.Chat, error) {
	chat, err := s.repo.CreateChat(ctx, user1, user2)
	if err != nil {
		return nil, err
	}

	s.log.Info("chat created", "chat_id", chat.ID)
	return chat, nil
}

func (s *ChatService) SendMessage(ctx context.Context, chatID int, sender, content string) error {
	return s.repo.SendMessage(ctx, chatID, sender, content)
}

func (s *ChatService) GetParticipants(ctx context.Context, chatID int) ([]string, error) {
	return s.repo.GetParticipants(ctx, chatID)
}

func (s *ChatService) GetMessages(ctx context.Context, chatID int, limit int) ([]models.Message, error) {
	return s.repo.GetMessages(ctx, chatID, limit)
}

func (s *ChatService) GetUserChats(ctx context.Context, username string) ([]models.ChatPreview, error) {
	return s.repo.GetUserChats(ctx, username)
}

func (s *ChatService) MarkChatRead(ctx context.Context, chatID int, username string) error {
	return s.repo.MarkChatRead(ctx, chatID, username)
}

func (s *ChatService) SetChatMatchTags(ctx context.Context, chatID int, tags []string) error {
	return s.repo.SetChatMatchTags(ctx, chatID, tags)
}
