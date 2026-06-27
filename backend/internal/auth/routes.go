package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/nicolas/finanzas/backend/internal/config"
)

// userIDFromToken adapts auth.Service.Me to extract the userID as a string.
// Kept as a closure so the routes file doesn't need to know the middleware type.
func userIDFromToken(svc Service) func(string) (string, error) {
	return func(token string) (string, error) {
		user, err := svc.Me(token)
		if err != nil {
			return "", err
		}
		return user.ID.String(), nil
	}
}

func RegisterRoutes(r *gin.RouterGroup, svc Service, cfg *config.Config) {
	h := NewHandler(svc, cfg)
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
		// /me uses the new middleware via direct call from /me route
		auth.GET("/me", meRoute(svc, h))
	}
}

func meRoute(svc Service, h *Handler) gin.HandlerFunc {
	// Hand-rolled: validate token via service, then call handler
	return func(c *gin.Context) {
		// Reuse RequireUserID logic inline
		uid, err := userIDFromToken(svc)(c.GetHeader("Authorization")[7:])
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": gin.H{"code": "INVALID_TOKEN"}})
			return
		}
		_ = uid
		h.Me(c)
	}
}