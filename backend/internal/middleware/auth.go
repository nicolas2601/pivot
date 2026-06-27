package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserIDResolver is the contract for middleware that resolves a token to a userID.
type UserIDResolver func(accessToken string) (userID string, err error)

// RequireUserID extracts Bearer token, resolves it to a userID, and stores
// the userID as "userID" in the gin context.
func RequireUserID(resolver UserIDResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "MISSING_TOKEN", "message": "Token requerido"}})
			return
		}
		userID, err := resolver(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": "Token inválido"}})
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}