package transport

import (
	"errors"
	"math"
	"net/http"
	listApp "xisain/mmomarket-backend/internal/modules/listing/application"
	"xisain/mmomarket-backend/internal/modules/order/application"
	db "xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type OrderHandler struct {
	service        *application.OrderService
	listingService *listApp.ListingService
}

func NewOrderHandler(
	service *application.OrderService,
	listingService *listApp.ListingService,
) *OrderHandler {
	return &OrderHandler{
		service:        service,
		listingService: listingService,
	}
}

type RegisterOrderRequest struct {
	ListingID string `json:"listing_id" binding:"required,uuid"`
	Quantity  int32  `json:"quantity" binding:"required,gt=0"`
}

func parseUUID(value string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	if err := uuid.Scan(value); err != nil {
		return uuid, err
	}
	return uuid, nil
}

func (o *OrderHandler) RegisterOrder(c *gin.Context) {
	value, exists := c.Get("authUser")
	user, ok := value.(db.User)
	if !exists || !ok || !user.ID.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":  http.StatusUnauthorized,
			"error": "unauthorized",
		})
		return
	}
	var req RegisterOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":  http.StatusBadRequest,
			"error": err.Error(),
		})
		return
	}
	listingID, err := parseUUID(req.ListingID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid listing_id",
		})
		return
	}
	listing, err := o.listingService.GetListingByID(c.Request.Context(), listingID)
	if errors.Is(err, listApp.ErrListingNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "listing not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	if !listing.Price.Valid || listing.Price.NaN || listing.Price.InfinityModifier != 0 || listing.Price.Int == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid listing price"})
		return
	}
	price, err := listing.Price.Int64Value()
	if err != nil || !price.Valid || price.Int64 < 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "listing price must be a non-negative integer within int64 range"})
		return
	}
	if price.Int64 > math.MaxInt64/int64(req.Quantity) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order total exceeds supported amount"})
		return
	}
	order, err := o.service.RegisterOrder(c.Request.Context(), application.RegisterOrderInput{
		BuyerID:     user.ID,
		SellerID:    listing.SellerID,
		ListingID:   listing.ID,
		Quantity:    req.Quantity,
		UnitPrice:   price.Int64,
		TotalAmount: price.Int64 * int64(req.Quantity),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":  http.StatusInternalServerError,
			"error": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  http.StatusOK,
		"order": order,
	})
}

func (o *OrderHandler) GetUserOrder(c *gin.Context) {
	value, exists := c.Get("authUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := value.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	orders, err := o.service.BuyerGetOrder(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (o *OrderHandler) UserUpdateOrder(c *gin.Context) {
	orderID, err := parseUUID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	value, exists := c.Get("authUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	user, ok := value.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	findOrder, err := o.service.BuyerFindOrderByID(c.Request.Context(), db.BuyerGetOrderByIDParams{
		ID:      orderID,
		BuyerID: user.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if findOrder.Status != "pending_payment" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order status is not pending"})
		return
	}
	updateOrder, err := o.service.BuyerUpdateOrder(c.Request.Context(), db.UserUpdateOrderParams{
		ID:      findOrder.ID,
		BuyerID: user.ID,
		Status:  "paid",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order": updateOrder})
}
