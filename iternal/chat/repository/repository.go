package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/chimort/course_project2/iternal/chat/models"
	"github.com/lib/pq"
)

type ChatRepository struct {
	db *sql.DB
}

var (
	ErrUsersBlocked    = errors.New("users are blocked")
	ErrCannotBlockSelf = errors.New("cannot block self")
)

func NewChatRepository(db *sql.DB) *ChatRepository {
	return &ChatRepository{db: db}
}

func (r *ChatRepository) CreateChat(ctx context.Context, user1, user2 string) (*models.Chat, error) {
	blocked, _, _, err := r.GetBlockStatus(ctx, user1, user2)
	if err != nil {
		return nil, err
	}
	if blocked {
		return nil, ErrUsersBlocked
	}

	tx, err := r.db.BeginTx(ctx, nil)
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

	return &models.Chat{
		ID:    fmt.Sprintf("%d", chatID),
		User1: user1,
		User2: user2,
	}, nil
}

func (r *ChatRepository) SendMessage(ctx context.Context, chatID int, sender, content string) error {
	blocked, err := r.IsChatBlockedForUser(ctx, chatID, sender)
	if err != nil {
		return err
	}
	if blocked {
		return ErrUsersBlocked
	}

	_, err = r.db.ExecContext(
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

	_, err = r.db.ExecContext(
		ctx,
		`UPDATE chat_histories
		 SET last_message_at = now()
		 WHERE chat_id = $1`,
		chatID,
	)
	return err
}

func (r *ChatRepository) GetPeerUsername(ctx context.Context, chatID int, username string) (string, error) {
	var peer string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT username
		 FROM chat_participants
		 WHERE chat_id = $1 AND username <> $2
		 LIMIT 1`,
		chatID,
		username,
	).Scan(&peer)
	return peer, err
}

func (r *ChatRepository) GetParticipants(ctx context.Context, chatID int) ([]string, error) {
	rows, err := r.db.QueryContext(
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

	return users, rows.Err()
}

func (r *ChatRepository) GetMessages(ctx context.Context, chatID int, limit int) ([]models.Message, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.db.QueryContext(
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

	messages := make([]models.Message, 0, limit)
	for rows.Next() {
		var m models.Message
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

	return messages, rows.Err()
}

func (r *ChatRepository) GetUserChats(ctx context.Context, username string) ([]models.ChatPreview, error) {
	return r.GetUserChatsFiltered(ctx, username, "recent", "", "", "")
}

func (r *ChatRepository) GetUserChatsFiltered(ctx context.Context, username, sortBy, filterMode, filterTag, peerQuery string) ([]models.ChatPreview, error) {
	orderClause := "sort_at DESC, chat_id DESC"
	switch strings.ToLower(strings.TrimSpace(sortBy)) {
	case "oldest":
		orderClause = "sort_at ASC, chat_id ASC"
	case "peer_asc":
		orderClause = "peer_username ASC, sort_at DESC, chat_id DESC"
	case "peer_desc":
		orderClause = "peer_username DESC, sort_at DESC, chat_id DESC"
	}

	rows, err := r.db.QueryContext(
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
				EXISTS (
					SELECT 1
					FROM messages um
					WHERE um.chat_id = cp.chat_id
					  AND um.sender_name <> $1
					  AND um.created_at > COALESCE(cp.last_read_at, to_timestamp(0))
				) AS has_unread,
				COALESCE(
					(
						SELECT hist.match_tags[
							1 + floor(random() * array_length(hist.match_tags, 1))::int
						]
						FROM chat_histories hist
						WHERE hist.chat_id = cp.chat_id
						  AND array_length(hist.match_tags, 1) > 0
					),
					''
				) AS match_hint,
				COALESCE(hist.match_tags, ARRAY[]::text[]) AS match_tags,
				COALESCE(hist.search_mode, 'MATCH_MODE_DEFAULT') AS search_mode,
				row_number() OVER (
					PARTITION BY COALESCE(peer.username, '')
					ORDER BY COALESCE(last_msg.created_at, ch.created_at) DESC, cp.chat_id DESC
				) AS rn
			FROM chat_participants cp
			JOIN chats ch ON ch.id = cp.chat_id
			LEFT JOIN chat_histories hist ON hist.chat_id = cp.chat_id
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
		SELECT chat_id, peer_username, last_message, last_message_at, has_unread, match_hint, match_tags, search_mode
		FROM raw
		WHERE rn = 1
		  AND ($2 = '' OR search_mode = $2)
		  AND ($3 = '' OR EXISTS (
			SELECT 1
			FROM unnest(match_tags) AS tag
			WHERE lower(tag) = lower($3)
		  ))
		  AND ($4 = '' OR peer_username ILIKE '%' || $4 || '%')
		ORDER BY `+orderClause,
		username,
		filterMode,
		filterTag,
		peerQuery,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	chats := make([]models.ChatPreview, 0)
	for rows.Next() {
		var chat models.ChatPreview
		var peerUsername sql.NullString
		if err := rows.Scan(
			&chat.ChatID,
			&peerUsername,
			&chat.LastMessage,
			&chat.LastMessageAt,
			&chat.HasUnread,
			&chat.MatchHint,
			pq.Array(&chat.MatchTags),
			&chat.SearchMode,
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

	return chats, rows.Err()
}

func (r *ChatRepository) MarkChatRead(ctx context.Context, chatID int, username string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE chat_participants
		 SET last_read_at = now()
		 WHERE chat_id = $1 AND username = $2`,
		chatID,
		username,
	)
	return err
}

func (r *ChatRepository) SetChatMatchTags(ctx context.Context, chatID int, tags []string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE chat_histories
		 SET match_tags = $2
		 WHERE chat_id = $1`,
		chatID,
		pq.Array(tags),
	)
	return err
}

func (r *ChatRepository) SetChatMatchMetadata(ctx context.Context, chatID int, tags []string, searchMode string) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE chat_histories
		 SET match_tags = $2,
		     search_mode = $3
		 WHERE chat_id = $1`,
		chatID,
		pq.Array(tags),
		searchMode,
	)
	return err
}

func (r *ChatRepository) BlockUser(ctx context.Context, blocker, blocked string) error {
	if blocker == "" || blocked == "" {
		return nil
	}
	if blocker == blocked {
		return ErrCannotBlockSelf
	}

	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO blocked_users (blocker_username, blocked_username)
		 VALUES ($1, $2)
		 ON CONFLICT (blocker_username, blocked_username) DO NOTHING`,
		blocker,
		blocked,
	)
	return err
}

func (r *ChatRepository) UnblockUser(ctx context.Context, blocker, blocked string) error {
	_, err := r.db.ExecContext(
		ctx,
		`DELETE FROM blocked_users
		 WHERE blocker_username = $1 AND blocked_username = $2`,
		blocker,
		blocked,
	)
	return err
}

func (r *ChatRepository) GetBlockStatus(ctx context.Context, user1, user2 string) (bool, bool, bool, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT blocker_username, blocked_username
		 FROM blocked_users
		 WHERE (blocker_username = $1 AND blocked_username = $2)
		    OR (blocker_username = $2 AND blocked_username = $1)`,
		user1,
		user2,
	)
	if err != nil {
		return false, false, false, err
	}
	defer rows.Close()

	var byUser1 bool
	var byUser2 bool
	for rows.Next() {
		var blocker string
		var blocked string
		if err := rows.Scan(&blocker, &blocked); err != nil {
			return false, false, false, err
		}
		if blocker == user1 && blocked == user2 {
			byUser1 = true
		}
		if blocker == user2 && blocked == user1 {
			byUser2 = true
		}
	}

	return byUser1 || byUser2, byUser1, byUser2, rows.Err()
}

func (r *ChatRepository) IsChatBlockedForUser(ctx context.Context, chatID int, username string) (bool, error) {
	peer, err := r.GetPeerUsername(ctx, chatID, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	blocked, _, _, err := r.GetBlockStatus(ctx, username, peer)
	return blocked, err
}
