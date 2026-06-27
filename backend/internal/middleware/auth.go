package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserIDResolver is the contract for middleware that resolves a token to a userID.
type UserIDResolver func(accessToken string) (userID string, err error)

// RequireUserID extracts Bearer token, resolves it to a userID, and stores
// the userID as "userID" in the gin context.
func RequireUserID(resolver UserIDResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := extractBearerToken(c)
		if err != nil {
			code := "MISSING_TOKEN"
			if errors.Is(err, ErrMalformedAuthHeader) {
				code = "MALFORMED_TOKEN"
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": code, "message": "Token requerido o mal formado"},
			})
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

// extractBearerToken is a thin gin-aware wrapper around the public ExtractBearer
// helper. Kept private so middleware callers go through RequireUserID; external
// callers should use ExtractBearer directly with c.GetHeader("Authorization").
func extractBearerToken(c *gin.Context) (string, error) {
	return ExtractBearer(c.GetHeader("Authorization"))
}
