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
	OrderApp "xisain/mmomarket-backend/internal/modules/order/application"
	OrderTransport "xisain/mmomarket-backend/internal/modules/order/transport"
	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"
	sellerTransport "xisain/mmomarket-backend/internal/modules/seller/transport"
	"xisain/mmomarket-backend/internal/platform/config"
	"xisain/mmomarket-backend/internal/platform/database"
	dbqueries "xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/gin-contrib/cors"
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
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
	}))
	router.Use(cfg.Sentry)
	// router.

	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}
	// Declare Transport & Service
	queries := dbqueries.New(db)
	authService := application.NewAuthService(queries)
	sellerService := sellerApp.NewSellerService(db)
	sellerHandler := sellerTransport.NewSellerHandler(sellerService)
	gameService := gameApp.NewGameService(db)
	gameHandler := gameTransport.NewGameHandler(gameService)
	orderService := OrderApp.NewOrderService(db)
	listingService := listingApp.NewListingService(db)
	listingHandler := listingTransport.NewListingHandler(listingService)
	orderHandler := OrderTransport.NewOrderHandler(orderService,listingService)


	authHandler := identityTransport.NewAuthHandler(
		authService,
		sellerService,
		false,
	)

	api := router.Group("/api/v1")

	identityTransport.RegisterAuthRoutes(
		api,
		authHandler,
		authService,
	)


	sellerTransport.RegisterSellerRoutes(
		api,
		sellerHandler,
		sellerService,
		authService,
	)
	
	gameTransport.RegisterGameRoutes(
		api,
		gameHandler,
		gameService,
		authService,
	)

	listingTransport.RegisterListingRoutes(api, listingHandler, authService)
	
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
