package auth

import (
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
		auth.GET("/me", middleware.RequireAuth(authMeFunc(svc)), h.Me)
	}
}

// authMeFunc adapts Service.Me to the middleware.AuthFunc signature.
func authMeFunc(svc Service) middleware.AuthFunc {
	return func(token string) (any, error) {
		user, err := svc.Me(token)
		if err != nil {
			return nil, err
		}
		return user, nil
	}
}