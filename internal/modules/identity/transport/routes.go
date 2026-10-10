package transport

import (
	"xisain/mmomarket-backend/internal/modules/identity/application"

	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(
	api *gin.RouterGroup,
	handler *AuthHandler,
	service *application.AuthService,
) {
	auth := api.Group("/auth")

	auth.POST("/register", handler.Register)
	auth.POST("/login", handler.Login)

	protected := auth.Group("")
	protected.Use(RequireAuth(service))

	protected.GET("/me", handler.Me)
	protected.POST("/logout", handler.Logout)
}
