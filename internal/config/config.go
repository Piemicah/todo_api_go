package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseUrl string
	JWTSecret   string
	Port        string
	Domain      string // e.g. "example.com" in prod, "" for localhost
	Env         string // "production" | "development"
}

func Load() (*Config, error) {
	// Load .env file in development. In production, environment variables should be set by the hosting environment (e.g. Docker, Kubernetes, cloud provider).
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, relying on environment variables")
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}

	var config *Config = &Config{
		DatabaseUrl: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
		Domain:      os.Getenv("COOKIE_DOMAIN"), // leave empty for localhost dev
		Env:         env,
	}
	return config, nil
}