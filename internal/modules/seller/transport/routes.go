package transport

import (
	identityApp "xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"
	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"

	// orderApp "xisain/mmomarket-backend/internal/modules/order/application"
	listingApp "xisain/mmomarket-backend/internal/modules/listing/application"
	listingTransport "xisain/mmomarket-backend/internal/modules/listing/transport"
	orderTransport "xisain/mmomarket-backend/internal/modules/order/transport"

	"github.com/gin-gonic/gin"
)

func RegisterSellerRoutes(
	api *gin.RouterGroup,
	handler *SellerHandler,
	orderHandler *orderTransport.OrderHandler,
	listingHandler *listingTransport.ListingHandler,
	sellerService *sellerApp.SellerService,
	authService *identityApp.AuthService,
	listingService *listingApp.ListingService,
) {

	seller := api.Group("/seller")
	seller.Use(identityTransport.RequireAuth(authService))
	seller.GET("/profile", handler.Profile)
	seller.GET("/listing-categories", listingHandler.ListListingCategories)
	seller.POST("/listings", listingHandler.RegisterListing)
	seller.GET("/listings", listingHandler.ListSellerListings)
	seller.GET("orders", orderHandler.ListSellerOrders)
	seller.POST("/applications", handler.Register)

	admin := api.Group("/admin/sellers")
	admin.Use(
		identityTransport.RequireAuth(authService),
		RequireAdmin(sellerService),
	)

	admin.GET("", handler.ListSellers)
	admin.POST("/:id/approve", handler.Approve)
	admin.POST("/:id/reject", handler.Reject)
}