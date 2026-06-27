package budgets

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, requireAuth gin.HandlerFunc) {
	g := r.Group("/budgets")
	g.Use(requireAuth)
	{
		g.GET("", h.List)
		g.POST("", h.Create)
		g.GET("/:id", h.Get)
		g.PATCH("/:id", h.Update)
		g.DELETE("/:id", h.Delete)
	}
}