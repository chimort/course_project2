package chat_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/chimort/course_project2/iternal/chat/repository"
	chat "github.com/chimort/course_project2/iternal/chat/service"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres dbname=users sslmode=disable")
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	// Чистим таблицы перед тестом
	_, _ = db.Exec("DELETE FROM messages")
	_, _ = db.Exec("DELETE FROM chat_participants")
	_, _ = db.Exec("DELETE FROM chat_histories")
	_, _ = db.Exec("DELETE FROM chats")
	_, _ = db.Exec("DELETE FROM users")

	return db
}

func insertTestUsers(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
        INSERT INTO users(username, first_name, last_name, email, password_hash)
        VALUES ('alice', 'Alice', 'Smith', 'alice@example.com', 'hash1'),
               ('bob', 'Bob', 'Brown', 'bob@example.com', 'hash2')
    `)
	require.NoError(t, err)
}

func TestChatFullFlow(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	logg := slog.New(slog.NewTextHandler(os.Stdout, nil))
	repo := repository.NewChatRepository(db)
	svc := chat.NewChatService(repo, logg)
	insertTestUsers(t, db)

	ctx := context.Background()

	// Создаём чат
	chat, err := svc.CreateChat(ctx, "alice", "bob")
	require.NoError(t, err)
	require.NotEmpty(t, chat.ID)

	// Проверяем chat_participants
	rows, err := db.Query("SELECT username FROM chat_participants WHERE chat_id = $1", chat.ID)
	require.NoError(t, err)
	defer rows.Close()

	participants := map[string]bool{}
	for rows.Next() {
		var username string
		require.NoError(t, rows.Scan(&username))
		participants[username] = true
	}
	require.True(t, participants["alice"])
	require.True(t, participants["bob"])

	// Проверяем chat_histories
	var chatHistoryID int
	err = db.QueryRow("SELECT chat_id FROM chat_histories WHERE chat_id = $1", chat.ID).Scan(&chatHistoryID)
	require.NoError(t, err)
	require.Equal(t, chat.ID, fmt.Sprintf("%d", chatHistoryID))

	// Создаём сообщение
	_, err = db.Exec(`
        INSERT INTO messages(chat_id, sender_name, content)
        VALUES ($1, $2, $3)
    `, chat.ID, "alice", "Hello Bob!")
	require.NoError(t, err)

	// Проверяем сообщение
	var msgContent string
	err = db.QueryRow("SELECT content FROM messages WHERE chat_id = $1 AND sender_name = $2", chat.ID, "alice").Scan(&msgContent)
	require.NoError(t, err)
	require.Equal(t, "Hello Bob!", msgContent)
}
