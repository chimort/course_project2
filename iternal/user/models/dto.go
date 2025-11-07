package models

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponce struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

type ProfileRequest struct {
	Username  string      `json:"username"`
	Interests []Interests `json:"interests"`
	Language  Language    `json:"language"`
}

type ProfileResponce struct {
	Username  string      `json:"username"`
	Interests []Interests `json:"interests"`
	Language  Language    `json:"language"`
}

type ChatHistoryResponce struct {
	ChatID      int64  `json:"chat_id"`
	PeerID      int64  `json:"peer_id"`
	PeerName    string `json:"peer_name"`
	UnreadCount int    `json:"unread_count"`
	LastMessage string `json:"last_message_preview"`
	LastUpdated int64  `json:"last_updated"`
}
