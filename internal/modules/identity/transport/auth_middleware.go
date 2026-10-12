package transport

import (
	"errors"
	"net/http"

	"xisain/mmomarket-backend/internal/modules/identity/application"
	"github.com/gin-gonic/gin"
)

func RequireAuth(
	service *application.AuthService,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("mmo_session")
		if err != nil || token == "" {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"success": false,
					"message": "Authentication required",
				},
			)
			return
		}

		user, err := service.GetCurrentUser(
			c.Request.Context(),
			token,
		)

		if errors.Is(err, application.ErrUnauthorized) {
			c.AbortWithStatusJSON(
				http.StatusUnauthorized,
				gin.H{
					"success": false,
					"message": "Invalid or expired session",
				},
			)
			return
		}
		if errors.Is(err, application.ErrSuspended) {
			c.AbortWithStatusJSON(
				http.StatusForbidden,
				gin.H{
					"success": false,
					"message": "Account is suspended",
				},
			)
			return
		}

		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusInternalServerError,
				gin.H{
					"success": false,
					"message": "Internal server error",
				},
			)
			return
		}

		c.Set("authUser", user)
		c.Next()
	}
}

func RequireAdmin(
    service *application.AuthService,
) gin.HandlerFunc {
    return func(c *gin.Context) {
        user, ok := currentUser(c)

        if !ok {
            c.AbortWithStatusJSON(
                http.StatusUnauthorized,
                gin.H{
                    "message": "Authentication required",
                },
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
                gin.H{
                    "message": "Authorization check failed",
                    "error": err.Error(),
                },
            )
            return
        }

        if !allowed {
            c.AbortWithStatusJSON(
                http.StatusForbidden,
                gin.H{
                    "message": "Admin access required",
                },
            )
            return
        }

        c.Next()
    }
}
