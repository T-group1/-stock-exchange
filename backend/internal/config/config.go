package config

import (
	"net/url"
	"os"
	"strconv"
	"time"
)

// Config содержит конфигурацию приложения
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Logger      LoggerConfig
	JWT         JWTConfig
	SMTP        SMTPConfig
	FrontendURL string // <--- ДОБАВЛЕНО ДЛЯ ЭТАПА 7
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DatabaseConfig конфигурация БД
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoggerConfig конфигурация логгера
type LoggerConfig struct {
	Level string
}

// JWTConfig конфигурация JWT
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// SMTPConfig конфигурация почты
type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}

// Load загружает конфигурацию из переменных окружения
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 30*time.Second),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "stock_exchange"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "super-secret-key-change-me"),
			AccessTokenExpiry:  getDurationEnv("JWT_ACCESS_EXPIRY", 15*60*time.Second),
			RefreshTokenExpiry: getDurationEnv("JWT_REFRESH_EXPIRY", 7*24*60*60*time.Second),
		},
		SMTP: SMTPConfig{
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getEnv("SMTP_PORT", ""),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"), // <--- ДОБАВЛЕНО
	}
}

// DSN возвращает строку подключения к БД
func (c *DatabaseConfig) DSN() string {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   c.Host + ":" + c.Port,
		Path:   c.DBName,
	}

	q := u.Query()
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()

	return u.String()
}

// Addr возвращает адрес для HTTP сервера
func (c *ServerConfig) Addr() string {
	return c.Host + ":" + c.Port
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}
