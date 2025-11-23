package config

import (
	"errors"
	"os"
)

const (
	ModeDevelopment = "development"
	ModeProduction  = "production"
)

type Config struct {
	Mode       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Port       string
}

// Load reads environment variables and returns a Config.
// Required: MODE, DATABASE_HOST, DATABASE_PORT, DATABASE_USER, DATABASE_PASSWORD, DATABASE_NAME
func Load() (*Config, error) {
	mode := os.Getenv("MODE")
	if mode == "" {
		return nil, errors.New("MODE is required")
	}
	if mode != ModeDevelopment && mode != ModeProduction {
		return nil, errors.New("invalid MODE")
	}

	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASSWORD")
	name := os.Getenv("DATABASE_NAME")
	appPort := os.Getenv("PORT")
	if host == "" || port == "" || user == "" || pass == "" || name == "" || appPort == "" {
		return nil, errors.New("database connection variables are required (host, port, user, password, name, appPort)")
	}

	return &Config{Mode: mode, DBHost: host, DBPort: port, DBUser: user, DBPassword: pass, DBName: name, Port: appPort}, nil
}
