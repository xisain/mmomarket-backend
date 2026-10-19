package transport

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"xisain/mmomarket-backend/internal/platform/database/db"

	"xisain/mmomarket-backend/internal/modules/listing/application"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type ListingHandler struct {
	service *application.ListingService
}

func NewListingHandler(
	service *application.ListingService,
) *ListingHandler {
	return &ListingHandler{service: service}
}

type RegisterListingRequest struct {
	GameID       string  `json:"game_id" binding:"required,uuid"`
	CategoryID   string  `json:"category_id" binding:"required,uuid"`
	GameRegionID *string `json:"game_region_id"`
	ServerID     *string `json:"server_id"`

	Name        string `json:"name" binding:"required,min=3,max=255"`
	Slug        string `json:"slug" binding:"required,min=3,max=255"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`

	Price           string `json:"price" binding:"required"`
	StockType       string `json:"stock_type" binding:"required,oneof=finite unlimited"`
	StockQuantity   *int64 `json:"stock_quantity"`
	QuantityPerUnit int64  `json:"quantity_per_unit" binding:"required,min=1"`
}

// Helper UUID wajib
func parseUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID

	if err := id.Scan(value); err != nil {
		return pgtype.UUID{}, err
	}

	if !id.Valid {
		return pgtype.UUID{}, fmt.Errorf("invalid UUID")
	}

	return id, nil
}

// Helper UUID optional
func parseOptionalUUID(value *string) (pgtype.UUID, error) {
	if value == nil {
		return pgtype.UUID{}, nil
	}

	return parseUUID(*value)
}

// Harga NUMERIC(12,2): maks 10 digit integer, 2 desimal
var pricePattern = regexp.MustCompile(
	`^\d{1,10}(?:\.\d{1,2})?$`,
)

func parsePrice(value string) (pgtype.Numeric, error) {
	value = strings.TrimSpace(value)

	if !pricePattern.MatchString(value) {
		return pgtype.Numeric{}, fmt.Errorf("invalid price")
	}

	// Konversi ke satuan sen tanpa float64
	parts := strings.SplitN(value, ".", 2)

	fractional := "00"
	if len(parts) == 2 {
		fractional = parts[1]
		if len(fractional) == 1 {
			fractional += "0"
		}
	}

	coefficient := new(big.Int)

	if _, ok := coefficient.SetString(
		parts[0]+fractional, 10,
	); !ok {
		return pgtype.Numeric{}, fmt.Errorf("invalid price")
	}

	return pgtype.Numeric{
		Int:   coefficient,
		Exp:   -2,
		Valid: true,
	}, nil
}

func (h *ListingHandler) RegisterListing(c *gin.Context) {

	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authentication required"})
		return
	}
	userID := user.ID
	sellerID, err := parseUUID(c.Param("seller_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid seller ID"})
		return
	}

	// 3. Bind JSON
	var req RegisterListingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	// 4. Parse UUID
	gameID, err := parseUUID(req.GameID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid game ID",
		})
		return
	}

	categoryID, err := parseUUID(req.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid category ID",
		})
		return
	}

	regionID, err := parseOptionalUUID(req.GameRegionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid region ID",
		})
		return
	}

	serverID, err := parseOptionalUUID(req.ServerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid server ID",
		})
		return
	}

	// 5. Parse harga
	price, err := parsePrice(req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid price",
		})
		return
	}

	// 6. Parse stock
	var stock pgtype.Int8

	if req.StockQuantity != nil {
		stock = pgtype.Int8{
			Int64: *req.StockQuantity,
			Valid: true,
		}
	}

	// 7. Parse image URL
	imageURL := pgtype.Text{}

	if strings.TrimSpace(req.ImageURL) != "" {
		imageURL = pgtype.Text{
			String: strings.TrimSpace(req.ImageURL),
			Valid:  true,
		}
	}

	// 8. Panggil application service
	listing, err := h.service.RegisterListing(
		c.Request.Context(),
		application.RegisterListingInput{
			UserID:       userID,
			SellerID:     sellerID,
			GameID:       gameID,
			CategoryID:   categoryID,
			GameRegionID: regionID,
			ServerID:     serverID,

			Name:        req.Name,
			Slug:        req.Slug,
			Description: req.Description,
			ImageURL:    imageURL,

			Price:           price,
			StockType:       req.StockType,
			StockQuantity:   stock,
			QuantityPerUnit: req.QuantityPerUnit,
		},
	)

	if err != nil {
		respondError(c, err)
		return
	}

	// 9. Response
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Listing berhasil dibuat",
		"data":    listing,
	})
}
func currentUser(c *gin.Context) (db.User, bool) {
	value, exists := c.Get("authUser")
	user, ok := value.(db.User)
	return user, exists && ok && user.ID.Valid
}

type RegisterListingCategoryRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=100"`
	Slug        string `json:"slug" binding:"required,min=3,max=100"`
	Description string `json:"description"`
}

func (h *ListingHandler) RegisterListingCategory(c *gin.Context) {
	var req RegisterListingCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid listing category data"})
		return
	}
	category, err := h.service.RegisterListingCategory(c.Request.Context(), application.RegisterListingCategoryInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Kategori listing berhasil dibuat", "data": category})
}

func (h *ListingHandler) ListListingCategories(c *gin.Context) {
	categories, err := h.service.ListListingCategories(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	if categories == nil {
		categories = []db.ListingCategory{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": categories})
}

func (h *ListingHandler) ListSellerListings(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authentication required"})
		return
	}
	sellerID, err := parseUUID(c.Param("seller_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid seller ID"})
		return
	}
	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 32)
	if err != nil || limit <= 0 || limit > 100 {
		respondError(c, application.ErrListingInvalidPagination)
		return
	}
	offset, err := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 32)
	if err != nil || offset < 0 {
		respondError(c, application.ErrListingInvalidPagination)
		return
	}
	listings, err := h.service.ListSellerListings(c.Request.Context(), user.ID, db.ListListingsBySellerParams{SellerID: sellerID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		respondError(c, err)
		return
	}
	if listings == nil {
		listings = []db.Listing{}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": listings})
}

func (h *ListingHandler) ListListingByGameAndCategory(c *gin.Context) {
	gameIDStr := c.Param("gameID")
	categoryIDStr := c.Param("categoryID")

	gameID, err := parseUUID(gameIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Game ID tidak valid",
		})
		return
	}

	categoryID, err := parseUUID(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Category ID tidak valid",
		})
		return
	}

	limit := int32(10)
	offset := int32(0)

	if raw := c.Query("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Limit tidak valid",
			})
			return
		}
		limit = int32(value)
	}

	if raw := c.Query("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Offset tidak valid",
			})
			return
		}
		offset = int32(value)
	}

	params := db.ListListingsByGameAndCategoryParams{
		GameID:     gameID,
		CategoryID: categoryID,
		Limit:      limit,
		Offset:     offset,
	}

	listings, err := h.service.ListListingsByGameAndCategory(
		c.Request.Context(),
		params,
	)
	if err != nil {
		respondError(c, err)
		return
	}

	if listings == nil {
		listings = []db.Listing{}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listings,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *ListingHandler) ListingByGameSlugAndCategorySlug(c *gin.Context) {
	gameSlug := c.Param("gameSlug")
	categorySlug := c.Param("categorySlug")
	limit := int32(10)
	offset := int32(0)
	params := db.ListListingsByGameSlugAndCategoryParams{
		Slug:       gameSlug,
		Slug_2:     categorySlug,
		Limit:      limit,
		Offset:     offset,
	}
	listings, err := h.service.ListListingByGameSlugAndCategorySlug(c.Request.Context(), params)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listings,
	})
}
func respondError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "Failed to process listing request"
	switch {
	case errors.Is(err, application.ErrListingSellerNotAllowed):
		status = http.StatusForbidden
	case errors.Is(err, application.ErrListingCategorySlugTaken), errors.Is(err, application.ErrListingCategoryAlreadyExists), errors.Is(err, application.ErrListingAlreadyExists):
		status = http.StatusConflict
	case errors.Is(err, application.ErrListingNotFound), errors.Is(err, application.ErrListingCategoryNotFound):
		status = http.StatusNotFound
	case errors.Is(err, application.ErrListingCategoryInvalidName), errors.Is(err, application.ErrListingCategoryInvalidSlug),
		errors.Is(err, application.ErrListingInvalidName), errors.Is(err, application.ErrListingInvalidSlug),
		errors.Is(err, application.ErrListingInvalidPrice), errors.Is(err, application.ErrListingInvalidStock),
		errors.Is(err, application.ErrListingInvalidRegion), errors.Is(err, application.ErrListingInvalidServer),
		errors.Is(err, application.ErrListingInvalidReference), errors.Is(err, application.ErrListingInvalidPagination):
		status = http.StatusBadRequest
	}
	if status != http.StatusInternalServerError {
		message = err.Error()
	}
	c.JSON(status, gin.H{"success": false, "message": message})
}
