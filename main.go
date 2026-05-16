package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"b2b-go.local/config"
	"b2b-go.local/internal/handler"
	"b2b-go.local/internal/integration"
	"b2b-go.local/internal/middleware"
	"b2b-go.local/internal/repository"
	"b2b-go.local/internal/service"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		logrus.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logrus.Fatal("Cannot ping DB:", err)
	}

	repo := repository.NewPostgresRepo(db)
	cbrClient := integration.NewCBRClient()
	emailSender := integration.NewEmailSender(cfg.SMTP)

	svc := service.NewBankService(repo, cbrClient, emailSender, cfg.JWTSecret, cfg.HMACSecret)
	h := handler.NewHandler(svc)

	// шедулер просроченных платежей
	go func() {
		for {
			time.Sleep(12 * time.Hour)
			svc.ProcessOverduePayments()
		}
	}()

	r := mux.NewRouter()

	// Публичные
	r.HandleFunc("/register", h.Register).Methods("POST")
	r.HandleFunc("/login", h.Login).Methods("POST")

	// Защищённые
	auth := r.PathPrefix("/").Subrouter()
	auth.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	auth.HandleFunc("/accounts", h.CreateAccount).Methods("POST")
	auth.HandleFunc("/accounts", h.GetAccounts).Methods("GET")
	auth.HandleFunc("/transfer", h.Transfer).Methods("POST")
	auth.HandleFunc("/cards", h.IssueCard).Methods("POST")
	auth.HandleFunc("/cards", h.GetCards).Methods("GET")
	auth.HandleFunc("/credits", h.ApplyCredit).Methods("POST")
	auth.HandleFunc("/credits/{creditId}/schedule", h.CreditSchedule).Methods("GET")
	auth.HandleFunc("/analytics", h.Analytics).Methods("GET")
	auth.HandleFunc("/accounts/{accountId}/predict", h.PredictBalance).Methods("GET")

	logrus.Info("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
