package transport

import (
	"errors"
	"net/http"
	"xisain/mmomarket-backend/internal/modules/game/application"
	db "xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameHandler struct {
	service *application.GameService
}

func NewGameHandler(
	service *application.GameService,
) *GameHandler {
	return &GameHandler{service: service}
}

type RegisterGameRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=255"`
	Slug        string `json:"slug" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"required,min=3,max=255"`
	Publisher   string `json:"publisher" binding:"required,min=3,max=255"`
}

type RegisterGameRegionRequest struct {
	Name string `json:"name" binding:"required,min=3,max=255"`
	Slug string `json:"slug" binding:"required,min=3,max=255"`
}

func gameResponse(g db.Game) gin.H {
	description := ""
	if g.Description.Valid {
		description = g.Description.String
	}

	publisher := ""
	if g.Publisher.Valid {
		publisher = g.Publisher.String
	}

	return gin.H{
		"id":          g.ID,
		"name":        g.Name,
		"slug":        g.Slug,
		"description": description,
		"publisher":   publisher,
		"is_active":   g.IsActive,
		"created_at":  g.CreatedAt,
		"updated_at":  g.UpdatedAt,
	}
}
func gameRegionResponse(g db.GameRegion) gin.H {
	return gin.H{
		"id":         g.ID,
		"game_id":    g.GameID,
		"name":       g.Name,
		"slug":       g.Slug,
		"is_active":  g.IsActive,
		"created_at": g.CreatedAt,
		"updated_at": g.UpdatedAt,
	}
}

type GameRegionListResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Game    gin.H   `json:"game"`
	Data    []gin.H `json:"data"`
}

type GameDetailResponse struct {
	ID          any    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Publisher   string `json:"publisher"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   any    `json:"created_at"`
	UpdatedAt   any    `json:"updated_at"`

	Categories any `json:"categories"`
	Regions    any `json:"regions"`
}

func parseUUID(value string) (pgtype.UUID, error) {
	var id pgtype.UUID
	err := id.Scan(value)
	return id, err
}
func (h *GameHandler) RegisterGame(c *gin.Context) {
	var req RegisterGameRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid game data",
		})
		return
	}
	game, err := h.service.RegisterGame(c.Request.Context(), application.RegisterGameInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Publisher:   req.Publisher,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create game",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Game created successfully",
		"data":    gameResponse(game),
	})
}

func (h *GameHandler) RegisterGameRegion(c *gin.Context) {
	// 1. Parse Game ID dari URL
	gameID, err := parseUUID(c.Param("game_id"))

	if err != nil || !gameID.Valid {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid game ID",
		})
		return
	}

	// 2. Validasi JSON body
	var req RegisterGameRegionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid game region data",
		})
		return
	}

	// 3. Panggil GameService
	region, err := h.service.RegisterRegionGame(
		c.Request.Context(),
		gameID,
		application.RegisterGameRegionInput{
			Name: req.Name,
			Slug: req.Slug,
		},
	)

	// 4. Error handling
	switch {
	case errors.Is(err, application.ErrGameNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Game not found",
		})
		return

	// case errors.Is(err, application.ErrGameInactive):
	//     c.JSON(http.StatusConflict, gin.H{
	//         "success": false,
	//         "message": "Game is inactive",
	//     })
	//     return

	case errors.Is(err, application.ErrRegionSlugTaken):
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "Region slug already exists",
		})
		return

	case errors.Is(err, application.ErrInvalidName),
		errors.Is(err, application.ErrInvalidSlug):
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return

	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create game region",
		})
		return
	}

	// 5. Response success
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Game region created successfully",
		"data":    gameRegionResponse(region),
	})
}
func (h *GameHandler) ListGame(c *gin.Context) {
	games, err := h.service.ListGames(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list games",
		})
		return
	}
	items := make([]gin.H, 0, len(games))
	for _, game := range games {
		items = append(items, gameResponse(game))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Game listed successfully",
		"data":    items,
	})

}
func (h *GameHandler) listRegionByGame(c *gin.Context) {
	gameID, err := parseUUID(c.Param("game_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid game ID",
		})
		return
	}
	game, err := h.service.GetGameByID(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get game",
		})
		return
	}
	gameRegs, err := h.service.ListRegionByGame(c.Request.Context(), gameID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list game regions",
		})
		return
	}
	items := make([]gin.H, 0, len(gameRegs))
	for _, gameReg := range gameRegs {
		items = append(items, gameRegionResponse(gameReg))
	}
	c.JSON(http.StatusOK, GameRegionListResponse{
		Success: true,
		Message: "Game regions listed successfully",
		Game:    gameResponse(game),
		Data:    items,
	})
}

func (h *GameHandler) ListGameActive(c *gin.Context) {
	games, err := h.service.ListGameActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list active games",
		})
		return
	}
	items := make([]gin.H, 0, len(games))
	for _, game := range games {
		items = append(items, gameResponse(game))
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}

func (h *GameHandler) GetGame(c *gin.Context) {
	slug := c.Param("slug")

	game, err := h.service.GetGame(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to get game",
		})
		return
	}

	gameCat, err := h.service.GetCategoriesByGame(
		c.Request.Context(),
		slug,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list game categories",
		})
		return
	}

	gameReg, err := h.service.ListRegionByGame(
		c.Request.Context(),
		game.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list game regions",
		})
		return
	}

	categories := make([]gin.H, 0, len(gameCat))

	for _, category := range gameCat {
		categories = append(categories, gin.H{
			"id":   category.CategoryID,
			"name": category.CategoryName,
			"slug": category.CategorySlug,
		})
	}

	regions := make([]gin.H, 0, len(gameReg))

	for _, region := range gameReg {
		regions = append(regions, gameRegionResponse(region))
	}

	// data := gameResponse(game)
	data := GameDetailResponse{
		ID:          game.ID,
		Name:        game.Name,
		Slug:        game.Slug,
		Description: game.Description.String,
		Publisher:   game.Publisher.String,
		IsActive:    game.IsActive,
		CreatedAt:   game.CreatedAt,
		UpdatedAt:   game.UpdatedAt,
		Categories:  categories,
		Regions:     regions,
	}
	

	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}
func (h *GameHandler) getCategoriesByGame(c *gin.Context) {
	slug := c.Param("slug")
	gameCat, err := h.service.GetCategoriesByGame(
		c.Request.Context(),
		slug,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to list game categories",
		})
		return
	}
	// gameListing, err := h.service
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gameCat,
	})
}