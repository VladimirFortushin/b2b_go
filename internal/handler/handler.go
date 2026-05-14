package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"b2b-go.local/internal/models"
	"b2b-go.local/internal/service"
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

// Заглушки для остальных эндпоинтов
func (h *Handler) IssueCard(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
func (h *Handler) ApplyCredit(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
func (h *Handler) CreditSchedule(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
func (h *Handler) Analytics(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
