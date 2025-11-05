package models

type Account struct {
	ID        int64     `json:"id" db:"id"`
    Email     string    `json:"email" db:"email"`
    Password  string    `json:"-" db:"password_hash"`
}