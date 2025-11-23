package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/davidr/bids-auth-service/internal/api"
	"github.com/davidr/bids-auth-service/internal/cache"
	"github.com/davidr/bids-auth-service/internal/config"
	"github.com/davidr/bids-auth-service/internal/db"
)

const (
	ModeDevelopment = "development"
	ModeProduction  = "production"
)

func main() {
	// Load development override file BEFORE config parsing if MODE indicates development.
	mode := os.Getenv("MODE")
	if mode != ModeDevelopment && mode != ModeProduction {
	}
	if mode == ModeDevelopment {
		if err := godotenv.Load(".env.Dev"); err != nil {
			log.Fatalf("Failed to load .env.Dev: %v", err)
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	dsn := cfg.DSN()
	pool, err := db.Connect(dsn)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer func() {
		if err := pool.Close(); err != nil {
			log.Printf("db close error: %v", err)
		}
	}()

	if mode != ModeProduction {
		if err := db.Migrate(dsn, "migrations"); err != nil {
			log.Fatalf("migration error: %v", err)
		}
	}

	refreshStore, err := cache.NewRedisRefreshStore(cfg)
	if err != nil {
		log.Fatalf("refresh store error: %v", err)
	}

	r := api.NewRouter(pool, cfg, refreshStore)

	addr := ":" + cfg.Port
	log.Printf("starting server on %s (mode=%s)", addr, mode)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
