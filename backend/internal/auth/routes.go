package auth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, svc Service, cfg *config.Config) {
	h := NewHandler(svc, cfg)
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/refresh", h.Refresh)
		auth.POST("/logout", h.Logout)
		// /me uses the inline meRoute handler which extracts the bearer token,
		// resolves the user, and stores it under "user" so h.Me can read it.
		auth.GET("/me", meRoute(svc, h))
	}
}

// meRoute is an inline handler that:
//  1. Validates the Authorization header via middleware.ExtractBearer.
//  2. Resolves the access token to a *User via Service.Me.
//  3. Stores the user in the gin context under "user" so Handler.Me can read it.
//
// We avoid the global RequireUserID middleware here because /me needs the
// full *User (not just the userID) for Handler.Me to return the user payload.
func meRoute(svc Service, h *Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := middleware.ExtractBearer(c.GetHeader("Authorization"))
		if err != nil {
			code := "MISSING_TOKEN"
			if errors.Is(err, middleware.ErrMalformedAuthHeader) {
				code = "MALFORMED_TOKEN"
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": code, "message": "Token requerido o mal formado"}})
			return
		}
		user, err := svc.Me(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": gin.H{"code": "INVALID_TOKEN", "message": "Token inválido"}})
			return
		}
		c.Set("user", user)
		h.Me(c)
	}
}
