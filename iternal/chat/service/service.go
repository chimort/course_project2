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

type ChatPreview struct {
	ChatID        string
	PeerUsername  string
	LastMessage   string
	LastMessageAt string
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

func (s *ChatService) GetMessages(
	ctx context.Context,
	chatID int,
	limit int,
) ([]Message, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT sender_name, content,
		        to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS created_at
		 FROM (
			SELECT sender_name, content, created_at, id
			FROM messages
			WHERE chat_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		 ) m
		 ORDER BY created_at ASC`,
		chatID,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]Message, 0, limit)
	for rows.Next() {
		var m Message
		var sender sql.NullString
		if err := rows.Scan(&sender, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}

		if sender.Valid && sender.String != "" {
			m.Sender = sender.String
		} else {
			m.Sender = "unknown"
		}
		m.ChatID = chatID
		messages = append(messages, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (s *ChatService) GetUserChats(
	ctx context.Context,
	username string,
) ([]ChatPreview, error) {

	rows, err := s.db.QueryContext(
		ctx,
		`WITH raw AS (
			SELECT
				cp.chat_id,
				peer.username AS peer_username,
				COALESCE(last_msg.content, '') AS last_message,
				COALESCE(last_msg.created_at, ch.created_at) AS sort_at,
				COALESCE(
					to_char(last_msg.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'),
					to_char(ch.created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
				) AS last_message_at,
				row_number() OVER (
					PARTITION BY COALESCE(peer.username, '')
					ORDER BY COALESCE(last_msg.created_at, ch.created_at) DESC, cp.chat_id DESC
				) AS rn
			FROM chat_participants cp
			JOIN chats ch ON ch.id = cp.chat_id
			LEFT JOIN chat_participants peer
				ON peer.chat_id = cp.chat_id
				AND peer.username <> $1
			LEFT JOIN LATERAL (
				SELECT m.content, m.created_at
				FROM messages m
				WHERE m.chat_id = cp.chat_id
				ORDER BY m.created_at DESC, m.id DESC
				LIMIT 1
			) AS last_msg ON true
			WHERE cp.username = $1
		)
		SELECT chat_id, peer_username, last_message, last_message_at
		FROM raw
		WHERE rn = 1
		ORDER BY sort_at DESC, chat_id DESC`,
		username,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]ChatPreview, 0)
	for rows.Next() {
		var chat ChatPreview
		var peerUsername sql.NullString

		if err := rows.Scan(
			&chat.ChatID,
			&peerUsername,
			&chat.LastMessage,
			&chat.LastMessageAt,
		); err != nil {
			return nil, err
		}

		if peerUsername.Valid && peerUsername.String != "" {
			chat.PeerUsername = peerUsername.String
		} else {
			chat.PeerUsername = "Unknown"
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return chats, nil
}
