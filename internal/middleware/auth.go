package middleware

import (
	"strings"

	"github.com/Sarthak-Nagaria/ticket-system/internal/utils"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Context keys used to pass authenticated user info to handlers.
const (
	ContextUserID = "userID"
	ContextEmail  = "email"
)

// AuthRequired validates JWT from Authorization: Bearer <token>.
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))

		if header == "" {
			utils.ErrorResponse(c, http.StatusUnauthorized, "missing Authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") ||
			strings.TrimSpace(parts[1]) == "" {

			utils.ErrorResponse(c, http.StatusUnauthorized, "Authorization header must be in 'Bearer <token>' format")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(strings.TrimSpace(parts[1]), jwtSecret)
		if err != nil {
			utils.ErrorResponse(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextEmail, claims.Email)

		c.Next()
	}
}