package models

type Account struct {
	ID        int     `json:"id"`
	UserID    int     `json:"user_id"`
	Balance   float64 `json:"balance"`
	CreatedAt string  `json:"created_at"`
}

type CreateAccountRequest struct {
	// пока пустой, счет создается для авторизованного пользователя
}
