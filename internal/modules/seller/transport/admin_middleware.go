package transport

import (
	"net/http"

	sellerApp "xisain/mmomarket-backend/internal/modules/seller/application"

	"github.com/gin-gonic/gin"
)

func RequireAdmin(
	service *sellerApp.SellerService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := currentUser(c)
		if !ok {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{"message": "Authentication required"},
			)
			return
		}

		allowed, err := service.IsAdmin(
			c.Request.Context(),
			user.ID,
		)
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{"message": "Authorization check failed"},
			)
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{"message": "Admin access required"},
			)
			return
		}

		c.Next()
	}
}