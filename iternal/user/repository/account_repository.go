package repository

import (
	"database/sql"
	"mathcing/iternal/user/models"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(account *models.Account) error {
	query := `insert into accounts (email, password_hash) values ($1, $2)`
	_, err := r.db.Exec(query, account.Email, account.Password)
	return err
}

func (r *AccountRepository) GetByEmail(email string) (*models.Account, error) {
	account := &models.Account{}
	query := `select id, email, password_hash from accounts where email = $1`
	err := r.db.QueryRow(query, email).Scan(&account.ID, &account.Email, &account.Password)
	if err != nil {
		return nil, err
	}
	return account, nil
}