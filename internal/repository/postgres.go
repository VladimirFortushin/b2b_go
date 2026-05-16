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
	CreateCard(card *models.Card) (int, error)
	GetCardsByUser(userID int) ([]models.Card, error)
	GetCardByID(cardID int) (*models.Card, error)
	ApplyCredit(credit *models.Credit) (int, error)
	CreatePaymentSchedule(creditID int, schedule []models.PaymentSchedule) error
	GetOverduePayments() ([]models.PaymentSchedule, error)
	MarkPaymentPaid(paymentID int) error
	UpdateAccountBalance(accountID int, amount float64) error
	AddTransaction(from, to int, amount float64) error
	GetMonthlyIncomeExpense(userID int, year, month int) (income, expense float64, err error)
	GetCreditLoad(userID int) (float64, error)
	GetFuturePayments(userID int, days int) ([]float64, error)
	GetPaymentSchedule(creditID int) ([]models.PaymentSchedule, error)
	GetCreditByID(creditID int) (*models.Credit, error)
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

func (r *postgresRepo) GetPaymentSchedule(creditID int) ([]models.PaymentSchedule, error) {
	rows, err := r.db.Query("SELECT id, credit_id, due_date, amount, paid FROM payment_schedules WHERE credit_id=$1 ORDER BY due_date", creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var schedule []models.PaymentSchedule
	for rows.Next() {
		var p models.PaymentSchedule
		if err := rows.Scan(&p.ID, &p.CreditID, &p.DueDate, &p.Amount, &p.Paid); err != nil {
			return nil, err
		}
		schedule = append(schedule, p)
	}
	return schedule, nil
}

func (r *postgresRepo) GetCreditByID(creditID int) (*models.Credit, error) {
	c := &models.Credit{}
	err := r.db.QueryRow("SELECT id, user_id, account_id, amount, rate, term_months, monthly_payment, status FROM credits WHERE id=$1", creditID).
		Scan(&c.ID, &c.UserID, &c.AccountID, &c.Amount, &c.Rate, &c.TermMonths, &c.MonthlyPayment, &c.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
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

	res, err := tx.Exec("UPDATE accounts SET balance = balance - $1 WHERE id=$2 AND balance >= $1", amount, fromID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("insufficient funds or account not found")
	}

	_, err = tx.Exec("UPDATE accounts SET balance = balance + $1 WHERE id=$2", amount, toID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO transactions (from_account, to_account, amount) VALUES ($1, $2, $3)", fromID, toID, amount)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *postgresRepo) CreateCard(card *models.Card) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO cards (account_id, card_number_enc, card_expiry_enc, cvv_hash, hmac, owner_id) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id",
		card.AccountID, card.CardNumberEnc, card.CardExpiryEnc, card.CVVHash, card.HMAC, card.OwnerID,
	).Scan(&id)
	return id, err
}

func (r *postgresRepo) GetCardsByUser(userID int) ([]models.Card, error) {
	rows, err := r.db.Query("SELECT id, account_id, card_number_enc, card_expiry_enc, cvv_hash, hmac, owner_id FROM cards WHERE owner_id=$1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.AccountID, &c.CardNumberEnc, &c.CardExpiryEnc, &c.CVVHash, &c.HMAC, &c.OwnerID); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *postgresRepo) GetCardByID(cardID int) (*models.Card, error) {
	c := &models.Card{}
	err := r.db.QueryRow("SELECT id, account_id, card_number_enc, card_expiry_enc, cvv_hash, hmac, owner_id FROM cards WHERE id=$1", cardID).
		Scan(&c.ID, &c.AccountID, &c.CardNumberEnc, &c.CardExpiryEnc, &c.CVVHash, &c.HMAC, &c.OwnerID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return c, err
}

func (r *postgresRepo) ApplyCredit(credit *models.Credit) (int, error) {
	var id int
	err := r.db.QueryRow(
		"INSERT INTO credits (user_id, account_id, amount, rate, term_months, monthly_payment, status) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id",
		credit.UserID, credit.AccountID, credit.Amount, credit.Rate, credit.TermMonths, credit.MonthlyPayment, credit.Status,
	).Scan(&id)
	return id, err
}

func (r *postgresRepo) CreatePaymentSchedule(creditID int, schedule []models.PaymentSchedule) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, p := range schedule {
		_, err := tx.Exec("INSERT INTO payment_schedules (credit_id, due_date, amount, paid) VALUES ($1,$2,$3,$4)",
			creditID, p.DueDate, p.Amount, false)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *postgresRepo) GetOverduePayments() ([]models.PaymentSchedule, error) {
	rows, err := r.db.Query("SELECT id, credit_id, due_date, amount, paid FROM payment_schedules WHERE paid = false AND due_date < CURRENT_DATE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var payments []models.PaymentSchedule
	for rows.Next() {
		var p models.PaymentSchedule
		if err := rows.Scan(&p.ID, &p.CreditID, &p.DueDate, &p.Amount, &p.Paid); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}
	return payments, nil
}

func (r *postgresRepo) MarkPaymentPaid(paymentID int) error {
	_, err := r.db.Exec("UPDATE payment_schedules SET paid = true WHERE id = $1", paymentID)
	return err
}

func (r *postgresRepo) UpdateAccountBalance(accountID int, amount float64) error {
	_, err := r.db.Exec("UPDATE accounts SET balance = balance + $1 WHERE id=$2", amount, accountID)
	return err
}

func (r *postgresRepo) AddTransaction(from, to int, amount float64) error {
	_, err := r.db.Exec("INSERT INTO transactions (from_account, to_account, amount) VALUES ($1, $2, $3)", from, to, amount)
	return err
}

func (r *postgresRepo) GetMonthlyIncomeExpense(userID int, year, month int) (income, expense float64, err error) {
	start := fmt.Sprintf("%d-%02d-01", year, month)
	end := fmt.Sprintf("%d-%02d-01", year, month+1)

	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(t.amount), 0) FROM transactions t
		JOIN accounts a ON t.to_account = a.id
		WHERE a.user_id = $1 AND t.created_at >= $2 AND t.created_at < $3`, userID, start, end).Scan(&income)
	if err != nil {
		return
	}

	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(t.amount), 0) FROM transactions t
		JOIN accounts a ON t.from_account = a.id
		WHERE a.user_id = $1 AND t.created_at >= $2 AND t.created_at < $3`, userID, start, end).Scan(&expense)
	return
}

func (r *postgresRepo) GetCreditLoad(userID int) (float64, error) {
	var load sql.NullFloat64
	err := r.db.QueryRow(`SELECT SUM(monthly_payment) FROM credits WHERE user_id=$1 AND status='active'`, userID).Scan(&load)
	if err != nil {
		return 0, err
	}
	if load.Valid {
		return load.Float64, nil
	}
	return 0, nil
}

func (r *postgresRepo) GetFuturePayments(userID int, days int) ([]float64, error) {
	query := `SELECT COALESCE(SUM(ps.amount), 0) FROM payment_schedules ps
    JOIN credits c ON ps.credit_id = c.id
    WHERE c.user_id = $1 AND ps.paid = false AND ps.due_date >= CURRENT_DATE AND ps.due_date <= CURRENT_DATE + ($2 * INTERVAL '1 day')`

	var total sql.NullFloat64
	err := r.db.QueryRow(query, userID, days).Scan(&total)
	if err != nil {
		return nil, err
	}
	res := []float64{0}
	if total.Valid {
		res[0] = total.Float64
	}
	return res, nil
}
