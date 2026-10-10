package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)
type Config struct {
	AppPort string
	DatabaseURL string
}
func Load() (*Config, error){
	_  = godotenv.Load()

	cfg := &Config {
		AppPort: os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if cfg.AppPort == "" {
		cfg.AppPort = "8000"
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg,nil
	
}