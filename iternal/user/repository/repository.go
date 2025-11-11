package repository

import (
	"context"
	"database/sql"

	"github.com/chimort/course_project2/iternal/user/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO users (username, first_name, last_name, email, password_hash, age, gender)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		user.Username, user.FirstName, user.LastName, user.Email, user.Password, user.Age, user.Gender,
	)
	if err != nil {
		return err
	}
	
	for _, l := range user.Languages {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO user_languages (username, language_id, proficiency_level)
			 VALUES ($1, (SELECT id FROM languages WHERE lang_name=$2), $3)`,
			user.Username, l.Language, l.Level,
		)
		if err != nil {
			return err
		}
	}

	for _, i := range user.Interests {
		_, err := tx.ExecContext(ctx,
			`INSERT INTO user_interests (username, interest_id)
			 VALUES ($1, (SELECT id FROM interests WHERE interest_name=$2))`,
			user.Username, i.Interest,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	u := &models.User{}

	err := r.db.QueryRowContext(ctx,
		`SELECT username, first_name, last_name, email, password_hash, age, gender
		 FROM users
		 WHERE username=$1`,
		username,
	).Scan(&u.Username, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.Age, &u.Gender)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT l.lang_name, ul.proficiency_level
		 FROM user_languages ul
		 JOIN languages l ON ul.language_id = l.id
		 WHERE ul.username=$1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	u.Languages = []models.UserLanguage{}
	for rows.Next() {
		var langName string
		var level string
		if err := rows.Scan(&langName, &level); err != nil {
			return nil, err
		}
		u.Languages = append(u.Languages, models.UserLanguage{
			Language: models.Language(langName),
			Level:    models.LanguageLevel(level),
		})
	}

	rows2, err := r.db.QueryContext(ctx,
		`SELECT i.interest_name
		 FROM user_interests ui
		 JOIN interests i ON ui.interest_id = i.id
		 WHERE ui.username=$1`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()

	u.Interests = []models.UserInterest{}
	for rows2.Next() {
		var interestName string
		if err := rows2.Scan(&interestName); err != nil {
			return nil, err
		}
		u.Interests = append(u.Interests, models.UserInterest{
			Interest: models.Interests(interestName),
		})
	}

	return u, nil
}
