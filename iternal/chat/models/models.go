package models

type Chat struct {
	ID    string
	User1 string
	User2 string
}

type Message struct {
	ID              int64
	ChatID          int
	Sender          string
	Content         string
	CreatedAt       string
	MessageType     string
	FileURL         string
	FileName        string
	MimeType        string
	FileSizeBytes   int64
	DurationSeconds int32
}

type ChatPreview struct {
	ChatID        string
	PeerUsername  string
	LastMessage   string
	LastMessageAt string
	HasUnread     bool
	MatchHint     string
	MatchTags     []string
	SearchMode    string
}
