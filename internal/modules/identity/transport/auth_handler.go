package transport

import (
	"errors"
	"fmt"
	"net/http"
	"xisain/mmomarket-backend/internal/modules/identity/application"
	sellapp "xisain/mmomarket-backend/internal/modules/seller/application"

	"xisain/mmomarket-backend/internal/platform/database/db"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	service       *application.AuthService
	sellerService *sellapp.SellerService
	secure        bool
}

func NewAuthHandler(
	service *application.AuthService,
	sellerService *sellapp.SellerService,
	secure bool,
) *AuthHandler {
	return &AuthHandler{
		service:       service,
		sellerService: sellerService,
		secure:        secure,
	}
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8,max=128"`
	DisplayName string `json:"display_name" binding:"required,min=2,max=150"`
}
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
type MeResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Seller      gin.H  `json:"seller"`
}

func publicUser(user db.User) gin.H {
	return gin.H{
		"id":           user.ID,
		"email":        user.Email,
		"display_name": user.DisplayName,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	user, err := h.service.Register(
		c.Request.Context(),
		req.Email,
		req.Password,
		req.DisplayName,
	)

	if errors.Is(err, application.ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"message": "Email already registered",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to register user",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Registration successful",
		"data":    publicUser(user),
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	fmt.Printf("Email: %s\n", req.Email)
	fmt.Printf("Password provided: %t\n", req.Password != "")

	user, token, err := h.service.Login(
		c.Request.Context(),
		req.Email,
		req.Password,
	)

	if errors.Is(err, application.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid email or password",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Login failed",
		})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)

	c.SetCookie(
		"mmo_session",
		token,
		7*24*60*60,
		"/",
		"",
		h.secure,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
		"data":    publicUser(user),
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := c.Get("authUser")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Authentication required",
		})
		return
	}
	authUser, ok := user.(db.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Invalid user data",
		})
		return
	}

	data := MeResponse{
		ID:          authUser.ID.String(),
		DisplayName: authUser.DisplayName,
		Email:       authUser.Email,
		Seller:      nil,
	}

	seller, err := h.sellerService.GetByOwner(
		c.Request.Context(),
		authUser.ID,
	)

	if err != nil {
		if !errors.Is(err, sellapp.ErrSellerNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to retrieve seller",
			})
			return
		}
	} else {
		if seller.ID.Valid {
			data.Seller = gin.H{
				"id": seller.ID.String(),
				"shop_name": seller.ShopName,
				"slug": seller.Slug,
				"status": seller.Status,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	token, _ := c.Cookie("mmo_session")

	if err := h.service.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Logout failed",
		})
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("mmo_session", "", -1, "/", "", h.secure, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}
func currentUser(c *gin.Context) (db.User, bool) {
	value, exists := c.Get("authUser")
	if !exists {
		return db.User{}, false
	}

	user, ok := value.(db.User)
	return user, ok
}
