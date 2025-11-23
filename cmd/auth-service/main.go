package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/davidr/bids-auth-service/internal/api"
	"github.com/davidr/bids-auth-service/internal/config"
	"github.com/davidr/bids-auth-service/internal/db"
)

func main() {
	// Load .env.Dev explicitly if present and MODE=development
	if os.Getenv("MODE") == config.ModeDevelopment {
		if err := godotenv.Load(".env.Dev"); err != nil {
			log.Fatalf("Error loading .env.Dev file")
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	dbConnStr := db.DSN(cfg)
	pool, err := db.Connect(dbConnStr)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	if cfg.Mode != config.ModeProduction {
		if err := db.Migrate(dbConnStr, "migrations"); err != nil {
			log.Fatalf("migration error: %v", err)
		}
	}

	r := api.NewRouter(pool)

	addr := ":" + cfg.Port
	log.Printf("starting server on %s (mode=%s)", addr, cfg.Mode)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
