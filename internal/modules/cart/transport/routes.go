package transport

import (
	identityApp "xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"

	"github.com/gin-gonic/gin"
)

func RegisterCartRoute(api *gin.RouterGroup, handler *CartHandler, authService *identityApp.AuthService) {
	cart := api.Group("/cart/me")
	cart.Use(identityTransport.RequireAuth(authService))
	cart.GET("", handler.getCart)
}
