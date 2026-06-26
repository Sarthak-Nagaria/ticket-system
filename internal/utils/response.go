package utils

import "github.com/gin-gonic/gin"

// ErrorResponse sends a consistent JSON error payload: {"error": "message"}.
func ErrorResponse(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// SuccessResponse sends a consistent JSON success payload by merging the
// provided data map into the response body.
func SuccessResponse(c *gin.Context, status int, data gin.H) {
	c.JSON(status, data)
}
