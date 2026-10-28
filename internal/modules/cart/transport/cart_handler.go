package transport

import (
	"net/http"
	"xisain/mmomarket-backend/internal/modules/cart/application"
	"xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	service *application.CartService
}

func NewCartHandler(
	service *application.CartService,
) *CartHandler {
	return &CartHandler{
		service: service,
	}
}

func (h *CartHandler) getCart(c *gin.Context) {
	value, exists := c.Get("authUser")

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	user, ok := value.(db.User)
	if !ok || !user.ID.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		return
	}

	cart, err := h.service.GetCart(c.Request.Context(),user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cart)



}
