package integration

import (
	"crypto/tls"
	"fmt"
	"log"

	"b2b-go.local/config"

	gomail "gopkg.in/gomail.v2"
)

type EmailSender struct {
	cfg config.SMTPConfig
}

func NewEmailSender(cfg config.SMTPConfig) *EmailSender {
	return &EmailSender{cfg: cfg}
}

func (s *EmailSender) SendPaymentEmail(to string, amount float64) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.User)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Платеж успешно проведен")
	body := fmt.Sprintf(`
		<h1>Спасибо за оплату!</h1>
		<p>Сумма: <strong>%.2f RUB</strong></p>
		<small>Это автоматическое уведомление</small>
	`, amount)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.cfg.Host, s.cfg.Port, s.cfg.User, s.cfg.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         s.cfg.Host,
		InsecureSkipVerify: false,
	}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("SMTP error: %v", err)
		return fmt.Errorf("email sending failed")
	}
	return nil
}
