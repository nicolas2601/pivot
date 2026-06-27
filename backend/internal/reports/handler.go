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
	c.JSON(http.StatusOK, ByCategoryResponse{From: from, To: to, Breakdown: rows})
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
	c.JSON(http.StatusOK, ByAccountResponse{From: from, To: to, Breakdown: rows})
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
	c.JSON(http.StatusOK, MonthlyTrendResponse{From: from, To: to, Points: pts})
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