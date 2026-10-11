package transport

import (
	"errors"
	"net/http"

	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"

	"github.com/gin-gonic/gin"
)

func (h *SellerHandler) ListSellers(c *gin.Context) {
	sellers, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to load sellers",
		})
		return
	}

	items := make([]gin.H, 0, len(sellers))
	for _, seller := range sellers {
		items = append(items, sellerResponse(seller))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": items,
	})
}

func (h *SellerHandler) Approve(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid seller ID",
		})
		return
	}

	seller, err := h.service.Approve(c.Request.Context(), id)
	h.respondReview(c, seller, err)
}

func (h *SellerHandler) Reject(c *gin.Context) {
	id, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid seller ID",
		})
		return
	}

	seller, err := h.service.Reject(c.Request.Context(), id)
	h.respondReview(c, seller, err)
}

func (h *SellerHandler) respondReview(
	c *gin.Context,
	seller interface{},
	err error,
) {
	if errors.Is(err, sellerApp.ErrSellerNotFound) {
		c.JSON(http.StatusConflict, gin.H{
			"message": "Seller not found or no longer pending",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to review seller",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": seller,
	})
}