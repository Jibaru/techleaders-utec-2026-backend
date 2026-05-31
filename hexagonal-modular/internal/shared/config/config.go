package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string
	DB       DBConfig
	Mail     MailConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// MailConfig is optional. If Host is empty main.go picks the no-op Sender
// adapter instead of the SMTP one.
type MailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		HTTPPort: getenv("HTTP_PORT", "8080"),
		DB: DBConfig{
			Host:     getenv("DB_HOST", "localhost"),
			Port:     getenv("DB_PORT", "5432"),
			User:     mustenv("DB_USER"),
			Password: mustenv("DB_PASSWORD"),
			Name:     mustenv("DB_NAME"),
			SSLMode:  getenv("DB_SSLMODE", "disable"),
		},
		Mail: MailConfig{
			Host:     getenv("SMTP_HOST", ""),
			Port:     getenv("SMTP_PORT", "587"),
			User:     getenv("SMTP_USER", ""),
			Password: getenv("SMTP_PASSWORD", ""),
			From:     getenv("MAIL_FROM", "coffee-loyalty@example.com"),
		},
	}

	if missing := missingVars(); len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required env vars: %v", missing)
	}
	return cfg, nil
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

var missing []string

func getenv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func mustenv(key string) string {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		missing = append(missing, key)
		return ""
	}
	return v
}

func missingVars() []string {
	out := missing
	missing = nil
	return out
}
