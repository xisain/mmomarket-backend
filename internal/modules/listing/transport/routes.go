package transport

import (
	"github.com/gin-gonic/gin"
	identityApp "xisain/mmomarket-backend/internal/modules/identity/application"
	identityTransport "xisain/mmomarket-backend/internal/modules/identity/transport"
)

func RegisterListingRoutes(api *gin.RouterGroup, handler *ListingHandler, authService *identityApp.AuthService) {
	admin := api.Group("/admin/listing-categories")
	admin.Use(identityTransport.RequireAuth(authService), identityTransport.RequireAdmin(authService))
	admin.POST("", handler.RegisterListingCategory)
	admin.GET("", handler.ListListingCategories)

	seller := api.Group("/seller/:seller_id")
	seller.Use(identityTransport.RequireAuth(authService))
	seller.GET("/listing-categories", handler.ListListingCategories)
	seller.POST("/listings", handler.RegisterListing)
	seller.GET("/listings", handler.ListSellerListings)

	public := api.Group("/public")
	// public.GET("/listings/:gameID/:categoryID", handler.ListListingByGameAndCategory)
	public.GET("/listings/:gameSlug/:categorySlug", handler.ListingByGameSlugAndCategorySlug)
	
}
