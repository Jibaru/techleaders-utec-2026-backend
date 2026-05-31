package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPPort string
	DB       DBConfig
	Mail     MailConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// MailConfig is optional. If Host is empty the SMTP sender becomes a no-op.
type MailConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// RedisConfig is optional. If Addr is empty main.go picks the noop cache.
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// KafkaConfig is optional. If Brokers is empty the API falls back to the
// direct SMTP/noop mail sender (no queue).
type KafkaConfig struct {
	Brokers   []string
	MailTopic string
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
		Redis: RedisConfig{
			Addr:     getenv("REDIS_ADDR", ""),
			Password: getenv("REDIS_PASSWORD", ""),
			DB:       atoi(getenv("REDIS_DB", "0"), 0),
		},
		Kafka: KafkaConfig{
			Brokers:   splitCSV(getenv("KAFKA_BROKERS", "")),
			MailTopic: getenv("KAFKA_MAIL_TOPIC", "mail.outbox"),
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

func atoi(s string, fallback int) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}

// splitCSV splits "a, b ,c" into ["a", "b", "c"]. Empty input → nil.
func splitCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
