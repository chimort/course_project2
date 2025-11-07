package models

type Profile struct {
	AccountID int64       `json:"account_id" db:"account_id"`
	Username  string      `json:"username" db:"username"`
	Interests []Interests `json:"interests" db:"interests"`
	Language  Language
}
