package models

type Credit struct {
	ID             int     `json:"id"`
	UserID         int     `json:"user_id"`
	AccountID      int     `json:"account_id"`
	Amount         float64 `json:"amount"`
	Rate           float64 `json:"rate"`
	TermMonths     int     `json:"term_months"`
	MonthlyPayment float64 `json:"monthly_payment"`
	Status         string  `json:"status"`
}

type PaymentSchedule struct {
	ID       int     `json:"id"`
	CreditID int     `json:"credit_id"`
	DueDate  string  `json:"due_date"`
	Amount   float64 `json:"amount"`
	Paid     bool    `json:"paid"`
}

type ApplyCreditRequest struct {
	AccountID int     `json:"account_id"`
	Amount    float64 `json:"amount"`
	Months    int     `json:"months"`
}
