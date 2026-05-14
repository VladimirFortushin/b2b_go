package service

import (
	"errors"
	"fmt"
	"time"

	"b2b-go.local/internal/models"
	"b2b-go.local/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type BankService struct {
	repo      repository.Repository
	jwtSecret []byte
}

func NewBankService(repo repository.Repository, jwtSecret string) *BankService {
	return &BankService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *BankService) Register(req models.RegisterRequest) (int, error) {
	// Проверка уникальности username
	existing, _ := s.repo.GetUserByUsername(req.Username)
	if existing != nil {
		return 0, errors.New("username already taken")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hashed),
	}
	return s.repo.CreateUser(user)
}

func (s *BankService) Authenticate(req models.LoginRequest) (string, error) {
	user, err := s.repo.GetUserByUsername(req.Username)
	if err != nil || user == nil {
		return "", errors.New("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d", user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *BankService) CreateAccount(userID int) (int, error) {
	acc := &models.Account{UserID: userID, Balance: 0}
	return s.repo.CreateAccount(acc)
}

func (s *BankService) GetAccounts(userID int) ([]models.Account, error) {
	return s.repo.GetAccountsByUser(userID)
}

func (s *BankService) Transfer(senderID int, fromAcc, toAcc int, amount float64) error {
	// Проверка, что fromAcc принадлежит senderID (упрощённо)
	accounts, err := s.repo.GetAccountsByUser(senderID)
	if err != nil {
		return err
	}
	belongs := false
	for _, a := range accounts {
		if a.ID == fromAcc {
			belongs = true
			break
		}
	}
	if !belongs {
		return errors.New("account does not belong to user")
	}

	return s.repo.TransferMoney(fromAcc, toAcc, amount)
}

// Заглушки для оставшихся фич (чтобы программа компилировалась)
func (s *BankService) IssueCard(userID int) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func (s *BankService) ApplyCredit(userID int, amount float64, months int) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func (s *BankService) GetCreditSchedule(creditID int) (interface{}, error) {
	return nil, errors.New("not implemented")
}
func (s *BankService) GetAnalytics(userID int) (interface{}, error) {
	return nil, errors.New("not implemented")
}
