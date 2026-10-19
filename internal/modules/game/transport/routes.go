package transport

import (
	"xisain/mmomarket-backend/internal/modules/game/application"
	identityApp "xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"

	"github.com/gin-gonic/gin"
)

func RegisterGameRoutes(api *gin.RouterGroup, handler *GameHandler, gameService *application.GameService, authService *identityApp.AuthService) {
	admin := api.Group("/admin/game")
	admin.Use(identityTransport.RequireAuth(authService), identityTransport.RequireAdmin(authService))

	admin.POST("", handler.RegisterGame)
	admin.GET("", handler.ListGame)
	admin.POST("/:game_id/", handler.RegisterGameRegion)
	admin.GET("/:game_id/", handler.listRegionByGame)
}
