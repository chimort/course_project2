package chat

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

)

type Chat struct {
	ID    string
	User1 string
	User2 string
}

type ChatService struct {
	log *slog.Logger
	db *sql.DB
}

func NewChatService(log *slog.Logger, db *sql.DB) *ChatService {
	return &ChatService{
		log: log.With("component", "chat_service"),
		db: db,
	}
}

func (s *ChatService) CreateChat(ctx context.Context, user1, user2 string) (*Chat, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var chatID int

	err = tx.QueryRowContext(ctx,
		`INSERT INTO chats (chat_type) VALUES ('private') RETURNING id`,
	).Scan(&chatID)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO chat_participants (chat_id, username) VALUES ($1, $2), ($1, $3)`,
		chatID, user1, user2,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO chat_histories (chat_id) VALUES ($1)`,
		chatID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	s.log.Info("chat created", "chat_id", chatID)

	return &Chat{
		ID:    fmt.Sprintf("%d", chatID),
		User1: user1,
		User2: user2,
	}, nil
}