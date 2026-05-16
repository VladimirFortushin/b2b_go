package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"b2b-go.local/internal/integration"
	"b2b-go.local/internal/models"
	"b2b-go.local/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type BankService struct {
	repo        repository.Repository
	cbrClient   *integration.CBRClient
	emailSender *integration.EmailSender
	jwtSecret   []byte
	hmacSecret  []byte
	encryptKey  []byte
	logger      *logrus.Logger
}

func NewBankService(repo repository.Repository, cbr *integration.CBRClient, email *integration.EmailSender, jwtSecret, hmacSecret string) *BankService {
	key := sha256.Sum256([]byte("bank-pgp-key-1234567890123456")) // фиксированный ключ для AES
	return &BankService{
		repo:        repo,
		cbrClient:   cbr,
		emailSender: email,
		jwtSecret:   []byte(jwtSecret),
		hmacSecret:  []byte(hmacSecret),
		encryptKey:  key[:],
		logger:      logrus.New(),
	}
}

func (s *BankService) Register(req models.RegisterRequest) (int, error) {
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

func (s *BankService) IssueCard(userID, accountID int) (*models.Card, error) {
	accs, err := s.repo.GetAccountsByUser(userID)
	if err != nil {
		return nil, err
	}
	found := false
	for _, a := range accs {
		if a.ID == accountID {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New("account not found or not yours")
	}

	number := generateLuhnCardNumber()
	expiry := time.Now().AddDate(3, 0, 0).Format("01/06") // ММ/ГГ
	cvv := fmt.Sprintf("%03d", randomInt(100, 999))

	encNumber, _ := s.encryptAES(number)
	encExpiry, _ := s.encryptAES(expiry)
	cvvHash, _ := bcrypt.GenerateFromPassword([]byte(cvv), bcrypt.DefaultCost)
	hmacValue := s.computeHMAC(number)

	card := &models.Card{
		AccountID:     accountID,
		CardNumberEnc: hex.EncodeToString(encNumber),
		CardExpiryEnc: hex.EncodeToString(encExpiry),
		CVVHash:       string(cvvHash),
		HMAC:          hmacValue,
		OwnerID:       userID,
	}
	id, err := s.repo.CreateCard(card)
	if err != nil {
		return nil, err
	}
	card.ID = id
	card.PlainNumber = number
	card.PlainExpiry = expiry
	return card, nil
}

func (s *BankService) GetCards(userID int) ([]models.Card, error) {
	cards, err := s.repo.GetCardsByUser(userID)
	if err != nil {
		return nil, err
	}

	for i := range cards {
		if cards[i].OwnerID == userID {
			encNum, _ := hex.DecodeString(cards[i].CardNumberEnc)
			encExp, _ := hex.DecodeString(cards[i].CardExpiryEnc)
			decNum, _ := s.decryptAES(encNum)
			decExp, _ := s.decryptAES(encExp)
			cards[i].PlainNumber = decNum
			cards[i].PlainExpiry = decExp
		}
	}
	return cards, nil
}

func (s *BankService) ApplyCredit(userID int, req models.ApplyCreditRequest) (*models.Credit, error) {
	rate, err := s.cbrClient.GetCentralBankRate()
	if err != nil {
		s.logger.Errorf("CBR error: %v, using default rate 15%%", err)
		rate = 15.0
	}
	monthly := annuityPayment(req.Amount, rate, req.Months)
	credit := &models.Credit{
		UserID:         userID,
		AccountID:      req.AccountID,
		Amount:         req.Amount,
		Rate:           rate,
		TermMonths:     req.Months,
		MonthlyPayment: math.Round(monthly*100) / 100,
		Status:         "active",
	}
	creditID, err := s.repo.ApplyCredit(credit)
	if err != nil {
		return nil, err
	}
	credit.ID = creditID

	var schedule []models.PaymentSchedule
	for m := 1; m <= req.Months; m++ {
		dueDate := time.Now().AddDate(0, m, 0).Format("2006-01-02")
		schedule = append(schedule, models.PaymentSchedule{
			CreditID: creditID,
			DueDate:  dueDate,
			Amount:   credit.MonthlyPayment,
			Paid:     false,
		})
	}
	err = s.repo.CreatePaymentSchedule(creditID, schedule)
	if err != nil {
		return nil, err
	}

	_ = s.repo.UpdateAccountBalance(req.AccountID, req.Amount)

	s.logger.Infof("Credit issued: user %d, amount %.2f", userID, req.Amount)
	return credit, nil
}

func (s *BankService) GetCreditSchedule(creditID int) ([]models.PaymentSchedule, error) {
	return s.repo.GetPaymentSchedule(creditID)
}

func (s *BankService) GetAnalytics(userID int) (*models.Analytics, error) {
	now := time.Now()
	income, expense, err := s.repo.GetMonthlyIncomeExpense(userID, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}
	load, err := s.repo.GetCreditLoad(userID)
	if err != nil {
		return nil, err
	}
	forecast, err := s.repo.GetFuturePayments(userID, 30)
	if err != nil {
		return nil, err
	}
	return &models.Analytics{
		Income:          income,
		Expense:         expense,
		CreditLoad:      load,
		BalanceForecast: forecast,
	}, nil
}

func (s *BankService) ProcessOverduePayments() {
	payments, err := s.repo.GetOverduePayments()
	if err != nil {
		s.logger.Errorf("Overdue processing error: %v", err)
		return
	}
	for _, p := range payments {
		penalty := p.Amount * 0.10
		totalDue := p.Amount + penalty
		credit, err := s.repo.GetCreditByID(p.CreditID)
		if err != nil || credit == nil {
			continue
		}
		err = s.repo.UpdateAccountBalance(credit.AccountID, -totalDue)
		if err != nil {
			s.logger.Warnf("Failed to charge overdue payment %d: %v", p.ID, err)
			continue
		}
		_ = s.repo.MarkPaymentPaid(p.ID)
		s.emailSender.SendPaymentEmail("user@example.com", totalDue)
	}
	s.logger.Info("Overdue payments processed")
}

func (s *BankService) encryptAES(plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return ciphertext, nil
}

func (s *BankService) decryptAES(ciphertext []byte) (string, error) {
	block, err := aes.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	return string(plaintext), err
}

func (s *BankService) computeHMAC(data string) string {
	h := hmac.New(sha256.New, s.hmacSecret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func generateLuhnCardNumber() string {
	prefix := "400000"
	acc := make([]byte, 16)
	copy(acc, prefix)
	for i := len(prefix); i < 15; i++ {
		acc[i] = byte('0' + randomInt(0, 9))
	}
	check := luhnChecksum(string(acc[:15]))
	acc[15] = byte('0' + check)
	return string(acc)
}

func luhnChecksum(num string) int {
	sum := 0
	for i, c := range num {
		digit, _ := strconv.Atoi(string(c))
		if i%2 == len(num)%2 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return (10 - (sum % 10)) % 10
}

func randomInt(min, max int) int {
	b := make([]byte, 1)
	rand.Read(b)
	return min + int(b[0])%(max-min+1)
}

func annuityPayment(principal float64, yearlyRate float64, months int) float64 {
	monthlyRate := yearlyRate / 12 / 100
	if monthlyRate == 0 {
		return principal / float64(months)
	}
	coeff := (monthlyRate * math.Pow(1+monthlyRate, float64(months))) / (math.Pow(1+monthlyRate, float64(months)) - 1)
	return principal * coeff
}
