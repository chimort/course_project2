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

type Message struct {
	ID        int64
	ChatID    int
	Sender    string
	Content   string
	CreatedAt string
}

type ChatService struct {
	log *slog.Logger
	db  *sql.DB
}

func NewChatService(log *slog.Logger, db *sql.DB) *ChatService {
	return &ChatService{
		log: log.With("component", "chat_service"),
		db:  db,
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
		`INSERT INTO chat_participants (chat_id, username) VALUES ($1,$2),($1,$3)`,
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

func (s *ChatService) SendMessage(
	ctx context.Context,
	chatID int,
	sender string,
	content string,
) error {

	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO messages (chat_id, sender_name, content)
		 VALUES ($1,$2,$3)`,
		chatID,
		sender,
		content,
	)

	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(
		ctx,
		`UPDATE chat_histories
		 SET last_message_at = now()
		 WHERE chat_id = $1`,
		chatID,
	)

	return err
}

func (s *ChatService) GetParticipants(
	ctx context.Context,
	chatID int,
) ([]string, error) {

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT username
		 FROM chat_participants
		 WHERE chat_id = $1`,
		chatID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string

	for rows.Next() {

		var u string

		if err := rows.Scan(&u); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	return users, nil
}