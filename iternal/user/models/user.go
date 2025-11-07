package models

type User struct {
	ID string `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Language []Language `db:"language"`
	Interests []Interests `db:"interests"`
}