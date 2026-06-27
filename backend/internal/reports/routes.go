package reports

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, requireAuth gin.HandlerFunc) {
	g := r.Group("/reports")
	g.Use(requireAuth)
	{
		g.GET("/by-category", h.ByCategory)
		g.GET("/by-account", h.ByAccount)
		g.GET("/monthly-trend", h.MonthlyTrend)
		g.GET("/budget-vs-actual", h.BudgetVsActual)
	}
}