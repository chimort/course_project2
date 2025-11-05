package models

type ChatHistory struct {
	ChatID       int64  `json:"chat_id"`
    PeerID       int64  `json:"peer_id"`
    PeerName     string `json:"peer_name"`
    UnreadCount  int    `json:"unread_count"`
    LastMessage  string `json:"last_message_preview"` 
    LastUpdated  int64  `json:"last_updated"`      
}