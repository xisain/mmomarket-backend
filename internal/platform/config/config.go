package config

import (
	"fmt"
	"os"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)
type Config struct {
	AppPort string
	DatabaseURL string
	Sentry gin.HandlerFunc
}
func Load() (*Config, error){
	_  = godotenv.Load()
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRYDSN"),
		EnableTracing: true,
		TracesSampleRate: 1,
		// EnableLogs: true,
		
	}); err != nil {
		fmt.Printf("sentry initilization failed %v", err)
	}
	cfg := &Config {
		AppPort: os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		Sentry: sentrygin.New(sentrygin.Options{
        Repanic: true,
    }),
	}
	if cfg.AppPort == "" {
		cfg.AppPort = "8000"
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg,nil
	
}