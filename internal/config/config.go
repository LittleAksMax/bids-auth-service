package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Port       string

	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	ValidationAPIKey   string

	RedisHost     string
	RedisPort     string
	RedisPassword string
}

// Load reads environment variables and returns a Config.
// Required: DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME, PORT,
// ACCESS_TOKEN_SECRET, REFRESH_TOKEN_SECRET, VALIDATION_API_KEY, REDIS_HOST, REDIS_PORT, REDIS_PASSWORD
func Load() (*Config, error) {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASSWORD")
	name := os.Getenv("DATABASE_NAME")
	appPort := os.Getenv("PORT")
	if host == "" || port == "" || user == "" || pass == "" || name == "" || appPort == "" {
		return nil, errors.New("database connection variables are required (host, port, user, password, name, PORT)")
	}

	accessSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	refreshSecret := os.Getenv("REFRESH_TOKEN_SECRET")
	validationKey := os.Getenv("VALIDATION_API_KEY")
	if accessSecret == "" || refreshSecret == "" || validationKey == "" {
		return nil, errors.New("ACCESS_TOKEN_SECRET, REFRESH_TOKEN_SECRET, VALIDATION_API_KEY are required")
	}

	accessTTL, err := parseDurationEnv("ACCESS_TOKEN_TTL", "15m")
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_TTL: %w", err)
	}
	refreshTTL, err := parseDurationEnv("REFRESH_TOKEN_TTL", "720h") // 30 days
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_TTL: %w", err)
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")

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
		RedisHost:          redisHost,
		RedisPort:          redisPort,
		RedisPassword:      redisPassword,
	}, nil
}

// DSN builds a Postgres connection string from component parts.
func (c *Config) DSN() string {
	userEsc := url.QueryEscape(c.DBUser)
	passEsc := url.QueryEscape(c.DBPassword)
	hostPort := net.JoinHostPort(c.DBHost, c.DBPort)
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", userEsc, passEsc, hostPort, c.DBName)
}

func parseDurationEnv(key, def string) (time.Duration, error) {
	val := os.Getenv(key)
	if val == "" {
		val = def
	}
	return time.ParseDuration(val)
}
