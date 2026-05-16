package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"b2b-go.local/internal/models"
	"b2b-go.local/internal/service"

	"github.com/gorilla/mux"
)

type Handler struct {
	svc *service.BankService
}

func NewHandler(svc *service.BankService) *Handler {
	return &Handler{svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Email == "" || len(req.Password) < 6 {
		http.Error(w, "validation failed", http.StatusBadRequest)
		return
	}
	id, err := h.svc.Register(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"user_id": id})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	token, err := h.svc.Authenticate(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	json.NewEncoder(w).Encode(models.LoginResponse{Token: token})
}

func (h *Handler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	accID, err := h.svc.CreateAccount(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"account_id": accID})
}

func (h *Handler) GetAccounts(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	accounts, err := h.svc.GetAccounts(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(accounts)
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	var req struct {
		FromAccountID int     `json:"from_account_id"`
		ToAccountID   int     `json:"to_account_id"`
		Amount        float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	if err := h.svc.Transfer(userID, req.FromAccountID, req.ToAccountID, req.Amount); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *Handler) IssueCard(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	var req models.IssueCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	card, err := h.svc.IssueCard(userID, req.AccountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

func (h *Handler) GetCards(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	cards, err := h.svc.GetCards(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(cards)
}

func (h *Handler) ApplyCredit(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	var req models.ApplyCreditRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	credit, err := h.svc.ApplyCredit(userID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(credit)
}

func (h *Handler) CreditSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	creditID, _ := strconv.Atoi(vars["creditId"])
	schedule, err := h.svc.GetCreditSchedule(creditID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(schedule)
}

func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	analytics, err := h.svc.GetAnalytics(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(analytics)
}

func (h *Handler) PredictBalance(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value("userID").(string))
	vars := mux.Vars(r)
	accountID, _ := strconv.Atoi(vars["accountId"])
	daysStr := r.URL.Query().Get("days")
	days, _ := strconv.Atoi(daysStr)
	if days > 365 {
		days = 365
	}

	analytics, err := h.svc.GetAnalytics(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_ = accountID
	json.NewEncoder(w).Encode(map[string]interface{}{
		"account_id": accountID,
		"days":       days,
		"forecast":   analytics.BalanceForecast,
	})
}
