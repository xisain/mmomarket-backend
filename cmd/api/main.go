package main

import (
	"context"
	// "fmt"
	"log"
	"net/http"
	"time"
	gameApp "xisain/mmomarket-backend/internal/modules/game/application"
	gameTransport "xisain/mmomarket-backend/internal/modules/game/transport"
	"xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"
	listingApp "xisain/mmomarket-backend/internal/modules/listing/application"
	listingTransport "xisain/mmomarket-backend/internal/modules/listing/transport"
	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"
	sellerTransport "xisain/mmomarket-backend/internal/modules/seller/transport"
	OrderApp "xisain/mmomarket-backend/internal/modules/order/application"
	OrderTransport "xisain/mmomarket-backend/internal/modules/order/transport"
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
	listingService := listingApp.NewListingService(db)
	listingHandler := listingTransport.NewListingHandler(listingService)
	listingTransport.RegisterListingRoutes(api, listingHandler, authService)
	orderService := OrderApp.NewOrderService(db)
	orderHandler := OrderTransport.NewOrderHandler(orderService,listingService)
	OrderTransport.RegisterOrderRoute(api, orderHandler, authService)
	
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
	// for _, route := range router.Routes() {
	// 	fmt.Printf("Method: %s | Path: %s | Handler: %s\n", route.Method, route.Path, route.Handler)
	// }
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
