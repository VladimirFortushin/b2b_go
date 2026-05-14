package repository

import (
	"database/sql"
	"fmt"

	"b2b-go.local/internal/models"
)

type Repository interface {
	CreateUser(user *models.User) (int, error)
	GetUserByUsername(username string) (*models.User, error)
	CreateAccount(account *models.Account) (int, error)
	GetAccountsByUser(userID int) ([]models.Account, error)
	TransferMoney(fromID, toID int, amount float64) error
	// можно добавить остальные методы позже
}

type postgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) Repository {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) CreateUser(user *models.User) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO users (username, email, password_hash) VALUES ($1, $2, $3) RETURNING id",
		user.Username, user.Email, user.PasswordHash,
	).Scan(&id)
	return id, err
}

func (r *postgresRepo) GetUserByUsername(username string) (*models.User, error) {
	u := &models.User{}
	err := r.db.QueryRow(
		"SELECT id, username, email, password_hash, created_at FROM users WHERE username=$1",
		username,
	).Scan(&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return u, err
}

func (r *postgresRepo) CreateAccount(account *models.Account) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO accounts (user_id, balance) VALUES ($1, $2) RETURNING id",
		account.UserID, account.Balance,
	).Scan(&id)
	return id, err
}

func (r *postgresRepo) GetAccountsByUser(userID int) ([]models.Account, error) {
	rows, err := r.db.Query("SELECT id, user_id, balance, created_at FROM accounts WHERE user_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []models.Account
	for rows.Next() {
		var a models.Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (r *postgresRepo) TransferMoney(fromID, toID int, amount float64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Списание
	res, err := tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id=$2 AND balance >= $1", amount, fromID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient funds or account not found")
	}

	// Зачисление
	_, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id=$2", amount, toID)
	if err != nil {
		return err
	}

	// Запись транзакции
	_, err = tx.Exec("INSERT INTO transactions (from_account, to_account, amount) VALUES ($1, $2, $3)", fromID, toID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}
