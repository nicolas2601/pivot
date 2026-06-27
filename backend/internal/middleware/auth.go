package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthFunc is the minimal contract middleware needs to validate a token.
// Caller (auth package) wraps its Service.Me to satisfy this.
type AuthFunc func(accessToken string) (user any, err error)

// RequireAuth extracts Bearer token, validates it via authn(token), and stores
// the user in the gin context as "user". Returns 401 on missing/invalid token.
func RequireAuth(authn AuthFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "MISSING_TOKEN", "message": "Token requerido"}})
			return
		}

		user, err := authn(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": "Token inválido"}})
			return
		}

		c.Set("user", user)
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