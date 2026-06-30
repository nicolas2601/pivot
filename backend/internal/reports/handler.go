package reports

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) ByCategory(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var q RangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	from, to, err := q.Resolved()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": err.Error()}})
		return
	}
	rows, err := h.svc.ByCategory(userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, ByCategoryResponse{From: from, To: to, Categories: rows})
}

func (h *Handler) ByAccount(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var q RangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	from, to, err := q.Resolved()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": err.Error()}})
		return
	}
	rows, err := h.svc.ByAccount(userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, ByAccountResponse{From: from, To: to, Accounts: rows})
}

func (h *Handler) MonthlyTrend(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var q RangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	from, to, err := q.Resolved()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": err.Error()}})
		return
	}
	pts, err := h.svc.MonthlyTrend(userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, MonthlyTrendResponse{From: from, To: to, Months: pts})
}

func (h *Handler) BudgetVsActual(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var q RangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	from, to, err := q.Resolved()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": err.Error()}})
		return
	}
	rows, err := h.svc.BudgetVsActual(userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, BudgetVsActualResponse{From: from, To: to, Rows: rows})
}

// Summary is the dashboard's headline numbers + per-day breakdown.
// Used by the front-end's /reports/summary schema.
func (h *Handler) Summary(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var q RangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	from, to, err := q.Resolved()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": err.Error()}})
		return
	}
	rep, err := h.svc.Summary(userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, rep)
}

// Cashflow returns income/expense totals plus the savings rate.
// Used by the front-end's /reports/cashflow schema.
func (h *Handler) Cashflow(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var q RangeQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	from, to, err := q.Resolved()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": err.Error()}})
		return
	}
	rep, err := h.svc.Cashflow(userID, from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, rep)
}