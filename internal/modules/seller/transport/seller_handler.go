package transport

import (
	"errors"
	"net/http"

	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"
	db "xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type SellerHandler struct {
	service *sellerApp.SellerService
}

func NewSellerHandler(
	service *sellerApp.SellerService,
) *SellerHandler {
	return &SellerHandler{service: service}
}

type RegisterSellerRequest struct {
	ShopName string `json:"shop_name" binding:"required,min=3,max=150"`
	Slug string `json:"slug" binding:"required,min=3,max=180"`
	Description string `json:"description" binding:"max=2000"`
}

func sellerResponse(s db.Seller) gin.H {
	return gin.H{
		"id":          s.ID,
		"owner_id":    s.OwnerID,
		"shop_name":   s.ShopName,
		"slug":        s.Slug,
		"description": s.Description,
		"status":      s.Status,
		"created_at":  s.CreatedAt,
	}
}

func currentUser(c *gin.Context) (db.User, bool) {
	value, exists := c.Get("authUser")
	if !exists {
		return db.User{}, false
	}

	user, ok := value.(db.User)
	return user, ok
}

func parseUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(value)
	return id, err
}

func (h *SellerHandler) Register(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Authentication required",
		})
		return
	}

	var req RegisterSellerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid seller data",
		})
		return
	}

	seller, err := h.service.Register(
		c.Request.Context(),
		user.ID,
		sellerApp.RegisterSellerInput{
			ShopName:    req.ShopName,
			Slug:        req.Slug,
			Description: req.Description,
		},
	)

	switch {
	case errors.Is(err, sellerApp.ErrSellerExists):
		c.JSON(http.StatusConflict, gin.H{
			"message": "You already have a store",
		})
	case errors.Is(err, sellerApp.ErrSlugTaken):
		c.JSON(http.StatusConflict, gin.H{
			"message": "Shop slug already taken",
		})
	case errors.Is(err, sellerApp.ErrInvalidSlug):
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid shop slug",
		})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to register seller",
		})
	default:
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "Seller application submitted",
			"data": sellerResponse(seller),
		})
	}
}

func (h *SellerHandler) Profile(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Authentication required",
		})
		return
	}

	seller, err := h.service.GetByOwner(
		c.Request.Context(), user.ID,
	)

	if errors.Is(err, sellerApp.ErrSellerNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Seller profile not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": sellerResponse(seller),
	})
}