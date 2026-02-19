package config

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/LittleAksMax/bids-util/env"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Port       int

	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	ValidationAPIKey   string
	TokenIssuer        string
	TokenAudience      string

	PasswordPepper string // Add this field for password pepper

	AllowedOrigins []string // CORS allowed origins, read from ALLOWED_ORIGINS (comma-separated)
}

// Load reads environment variables and returns a Config.
// Required: DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME, PORT,
// ACCESS_TOKEN_SECRET, REFRESH_TOKEN_SECRET, VALIDATION_API_KEY, REDIS_HOST, REDIS_PORT, REDIS_PASSWORD
func Load() (*Config, error) {
	host := env.GetStrFromEnv("DATABASE_HOST")
	port := env.GetStrFromEnv("DATABASE_PORT")
	user := env.GetStrFromEnv("DATABASE_USER")
	pass := env.GetStrFromEnv("DATABASE_PASSWORD")
	name := env.GetStrFromEnv("DATABASE_NAME")
	appPort := env.ReadPort("PORT")

	accessSecret := env.GetStrFromEnv("ACCESS_TOKEN_SECRET")
	refreshSecret := env.GetStrFromEnv("REFRESH_TOKEN_SECRET")
	validationKey := env.GetStrFromEnv("VALIDATION_API_KEY")
	pepper := env.GetStrFromEnv("PASSWORD_PEPPER")

	// Token settings
	accessTTL := env.ParseDurationEnv("ACCESS_TOKEN_TTL")
	refreshTTL := env.ParseDurationEnv("REFRESH_TOKEN_TTL")
	tokenIssuer := env.GetStrFromEnv("TOKEN_ISSUER")
	tokenAudience := env.GetStrFromEnv("TOKEN_AUDIENCE")

	// CORS settings
	allowedOrigins := env.GetStrListFromEnv("ALLOWED_ORIGINS")

	return &Config{
		DBHost:             host,
		DBPort:             port,
		DBUser:             user,
		DBPassword:         pass,
		DBName:             name,
		Port:               appPort,
		AccessTokenSecret:  accessSecret,
		RefreshTokenSecret: refreshSecret,
		ValidationAPIKey:   validationKey,
		AccessTokenTTL:     accessTTL,
		RefreshTokenTTL:    refreshTTL,
		TokenIssuer:        tokenIssuer,
		TokenAudience:      tokenAudience,
		PasswordPepper:     pepper,
		AllowedOrigins:     allowedOrigins,
	}, nil
}

// DSN builds a Postgres connection string from component parts.
func (c *Config) DSN() string {
	userEsc := url.QueryEscape(c.DBUser)
	passEsc := url.QueryEscape(c.DBPassword)
	hostPort := net.JoinHostPort(c.DBHost, c.DBPort)
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", userEsc, passEsc, hostPort, c.DBName)
}
