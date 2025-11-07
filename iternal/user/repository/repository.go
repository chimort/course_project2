package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/chimort/course_project2/iternal/user/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	langJSON, _ := json.Marshal(user.Language)
	interestsJSON, _ := json.Marshal(user.Interests)

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, username, password, language, interests)
		 VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Username, user.Password, langJSON, interestsJSON,
	)
	return err
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, username, password, language, interests FROM users WHERE username=$1`,
		username,
	)

	var u models.User
	var langJSON, interestsJSON []byte
	if err := row.Scan(&u.ID, &u.Username, &u.Password, &langJSON, &interestsJSON); err != nil {
		return nil, err
	}

	json.Unmarshal(langJSON, &u.Language)
	json.Unmarshal(interestsJSON, &u.Interests)

	return &u, nil
}
