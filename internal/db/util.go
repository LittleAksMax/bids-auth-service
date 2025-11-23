package db

import (
	"fmt"
	"net"
	"net/url"

	"github.com/davidr/bids-auth-service/internal/config"
)

// DSN constructs a Postgres connection string from environment variables.
// Required env vars: DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME.
func DSN(cfg *config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		url.QueryEscape(cfg.DBUser),
		url.QueryEscape(cfg.DBPassword),
		net.JoinHostPort(cfg.DBHost, cfg.DBPort),
		cfg.DBName,
	)
}
