package config

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret   string
	SMTP        SMTPConfig
}

type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost/bank?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "my-super-secret-key"),
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", "localhost"),
			Port:     587,
			User:     getEnv("SMTP_USER", "noreply@bank.local"),
			Password: getEnv("SMTP_PASS", "password"),
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
