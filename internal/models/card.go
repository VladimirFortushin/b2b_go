package models

type Card struct {
	ID            int    `json:"id"`
	AccountID     int    `json:"account_id"`
	CardNumberEnc string `json:"-"`
	CardExpiryEnc string `json:"-"`
	CVVHash       string `json:"-"`
	HMAC          string `json:"hmac"`
	OwnerID       int    `json:"owner_id"`
	PlainNumber   string `json:"card_number,omitempty"`
	PlainExpiry   string `json:"expiry,omitempty"`
}

type IssueCardRequest struct {
	AccountID int `json:"account_id"`
}
