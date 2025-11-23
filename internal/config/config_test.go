package config

import (
	"os"
	"strings"
	"testing"
)

func setBaseEnv() {
	os.Setenv("MODE", ModeTest)
	os.Setenv("DATABASE_HOST", "localhost")
	os.Setenv("DATABASE_PORT", "5432")
	os.Setenv("DATABASE_USER", "user")
	os.Setenv("DATABASE_PASSWORD", "p@ss word")
	os.Setenv("DATABASE_NAME", "db")
}

func TestLoadConfig(t *testing.T) {
	setBaseEnv()
	os.Setenv("PORT", "9000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Mode != ModeTest {
		t.Fatalf("expected mode test got %s", cfg.Mode)
	}
	if cfg.Port != "9000" {
		t.Fatalf("expected port 9000 got %s", cfg.Port)
	}
	dsn := cfg.DSN()
	if !strings.Contains(dsn, "postgres://user:p%40ss+word@localhost:5432/db") {
		t.Fatalf("dsn encoding mismatch: %s", dsn)
	}
}

func TestLoadConfigMissingDBVar(t *testing.T) {
	os.Clearenv()
	os.Setenv("MODE", ModeTest)
	os.Setenv("DATABASE_HOST", "localhost")
	// Missing port
	os.Setenv("DATABASE_USER", "user")
	os.Setenv("DATABASE_PASSWORD", "pass")
	os.Setenv("DATABASE_NAME", "db")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for missing vars")
	}
}

func TestLoadConfigInvalidMode(t *testing.T) {
	setBaseEnv()
	os.Setenv("MODE", "invalid")
	if _, err := Load(); err == nil {
		t.Fatalf("expected error for invalid mode")
	}
}
