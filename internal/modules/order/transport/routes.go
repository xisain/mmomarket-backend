package transport

import (
	identityApp "xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"

	"github.com/gin-gonic/gin"
)

func RegisterOrderRoute(api *gin.RouterGroup, handler *OrderHandler, authService *identityApp.AuthService) {
	order := api.Group("/orders")
	order.Use(identityTransport.RequireAuth(authService))
	order.POST("", handler.RegisterOrder)
	order.GET("", handler.GetUserOrder)
	order.PUT("/:id", handler.UserUpdateOrder)
}
