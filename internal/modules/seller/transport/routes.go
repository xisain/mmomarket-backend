package transport

import (
	identityApp "xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"
	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"

	"github.com/gin-gonic/gin"
)

func RegisterSellerRoutes(
	api *gin.RouterGroup,
	handler *SellerHandler,
	sellerService *sellerApp.SellerService,
	authService *identityApp.AuthService,
) {
	seller := api.Group("/seller")
	seller.Use(identityTransport.RequireAuth(authService))

	seller.POST("/applications", handler.Register)
	seller.GET("/profile", handler.Profile)

	admin := api.Group("/admin/sellers")
	admin.Use(
		identityTransport.RequireAuth(authService),
		RequireAdmin(sellerService),
	)

	admin.GET("", handler.ListSellers)
	admin.POST("/:id/approve", handler.Approve)
	admin.POST("/:id/reject", handler.Reject)
}