package accounts

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts account CRUD endpoints under /accounts.
// All routes require a valid Bearer token; the middleware extracts the
// userID and stores it in the gin context as "userID".
func RegisterRoutes(r *gin.RouterGroup, h *Handler, requireAuth gin.HandlerFunc) {
	g := r.Group("/accounts")
	g.Use(requireAuth)
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.PATCH("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}