package models

type Analytics struct {
	Income          float64   `json:"income"`
	Expense         float64   `json:"expense"`
	CreditLoad      float64   `json:"credit_load"`
	BalanceForecast []float64 `json:"balance_forecast"`
}
