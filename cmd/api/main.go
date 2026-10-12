package main

import (
	"context"
	"log"
	"net/http"
	"time"
	"xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"
	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"
	sellerTransport "xisain/mmomarket-backend/internal/modules/seller/transport"
	gameApp "xisain/mmomarket-backend/internal/modules/game/application"
	gameTransport "xisain/mmomarket-backend/internal/modules/game/transport"
	"xisain/mmomarket-backend/internal/platform/config"
	"xisain/mmomarket-backend/internal/platform/database"
	dbqueries "xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.Connect(
		context.Background(),
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(cfg.Sentry)
	// router.

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}

	queries := dbqueries.New(db)

	authService := application.NewAuthService(queries)

	authHandler := identityTransport.NewAuthHandler(
		authService,
		false,
	)

	api := router.Group("/api/v1")

	identityTransport.RegisterAuthRoutes(
		api,
		authHandler,
		authService,
	)
	sellerService := sellerApp.NewSellerService(db)
	sellerHandler := sellerTransport.NewSellerHandler(sellerService)

	sellerTransport.RegisterSellerRoutes(
		api,
		sellerHandler,
		sellerService,
		authService,
	)
	gameService := gameApp.NewGameService(db)
	gameHandler := gameTransport.NewGameHandler(gameService)
	gameTransport.RegisterGameRoutes(
		api,
		gameHandler,
		gameService,
		authService,
	)
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "MMO Marketplace API running",
		})
	})

	api.GET("/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"message": "Database unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Database connected",
		})
	})
	api.GET("/test-sentry", func(c *gin.Context) {
		panic("Test Sentry Error!")
	})

	log.Printf("Starting API on port %s", cfg.AppPort)

	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
