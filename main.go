package main

import (
	"database/sql"
	"log"
	"net/http"

	"b2b-go.local/config"
	"b2b-go.local/internal/handler"
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
	svc := service.NewBankService(repo, cfg.JWTSecret)
	h := handler.NewHandler(svc)

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
	auth.HandleFunc("/credits", h.ApplyCredit).Methods("POST")
	auth.HandleFunc("/credits/{creditId}/schedule", h.CreditSchedule).Methods("GET")
	auth.HandleFunc("/analytics", h.Analytics).Methods("GET")

	logrus.Info("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
