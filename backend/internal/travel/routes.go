package travel

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *Handler, requireAuth gin.HandlerFunc) {
	g := r.Group("/travel/groups")
	g.Use(requireAuth)
	{
		g.POST("", h.CreateGroup)
		g.GET("", h.ListGroups)

		// Per-group sub-resources
		g.GET("/:id", h.GetGroup)
		g.PATCH("/:id", h.UpdateGroup)
		g.DELETE("/:id", h.DeleteGroup)

		// Members
		g.POST("/:id/members", h.AddMember)
		g.GET("/:id/members", h.ListMembers)
		g.DELETE("/:id/members/:user_id", h.RemoveMember)

		// Expenses
		g.POST("/:id/expenses", h.AddExpense)
		g.GET("/:id/expenses", h.ListExpenses)
		g.GET("/:id/expenses/:expense_id", h.GetExpense)
		g.DELETE("/:id/expenses/:expense_id", h.DeleteExpense)

		// Settlements
		g.GET("/:id/settlements", h.GetSettlements)
		g.POST("/:id/settlements", h.RecordSettlement)
		g.GET("/:id/settlements/list", h.ListSettlements)
		g.POST("/:id/settlements/:settlement_id/confirm", h.ConfirmSettlement)
	}
}