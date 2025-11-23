package config

import (
	"os"
	"strings"
	"testing"
)

func setBaseEnv() {
	_ = os.Setenv("DATABASE_HOST", "localhost")
	_ = os.Setenv("DATABASE_PORT", "5432")
	_ = os.Setenv("DATABASE_USER", "user")
	_ = os.Setenv("DATABASE_PASSWORD", "p@ss word")
	_ = os.Setenv("DATABASE_NAME", "db")
	_ = os.Setenv("ACCESS_TOKEN_SECRET", "access")
	_ = os.Setenv("REFRESH_TOKEN_SECRET", "refresh")
	_ = os.Setenv("VALIDATION_API_KEY", "val-key")
	_ = os.Setenv("ACCESS_TOKEN_TTL", "10m")
	_ = os.Setenv("REFRESH_TOKEN_TTL", "24h")
}

func TestLoadConfig(t *testing.T) {
	os.Clearenv()
	setBaseEnv()
	_ = os.Setenv("PORT", "9000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9000" {
		t.Fatalf("expected port 9000 got %s", cfg.Port)
	}
	if cfg.AccessTokenTTL.Minutes() != 10 {
		t.Fatalf("access TTL not parsed")
	}
	if cfg.RefreshTokenTTL.Hours() != 24 {
		t.Fatalf("refresh TTL not parsed")
	}
	dsn := cfg.DSN()
	if !strings.Contains(dsn, "postgres://user:p%40ss+word@localhost:5432/db") {
		t.Fatalf("dsn encoding mismatch: %s", dsn)
	}
}

func TestLoadConfigMissingDBVar(t *testing.T) {
	os.Clearenv()
	_ = os.Setenv("DATABASE_HOST", "localhost")
	// Missing port
	_ = os.Setenv("DATABASE_USER", "user")
	_ = os.Setenv("DATABASE_PASSWORD", "pass")
	_ = os.Setenv("DATABASE_NAME", "db")
	_ = os.Setenv("PORT", "8000")
	_ = os.Setenv("ACCESS_TOKEN_SECRET", "a")
	_ = os.Setenv("REFRESH_TOKEN_SECRET", "b")
	_ = os.Setenv("VALIDATION_API_KEY", "k")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for missing vars")
	}
}
