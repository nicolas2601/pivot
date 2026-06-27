package budgets

import (
	"errors"
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

func (h *Handler) Create(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var dto CreateRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	b, err := h.svc.Create(userID, dto.ToServiceCreate())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, b)
}

func (h *Handler) List(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	list, err := h.svc.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Budgets: list})
}

func (h *Handler) Get(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	id, _ := uuid.Parse(c.Param("id"))
	b, err := h.svc.Get(id, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) Update(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	id, _ := uuid.Parse(c.Param("id"))
	var dto UpdateRequestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	b, err := h.svc.Update(id, userID, dto.ToServiceUpdate())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, b)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.Delete(id, userID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrBudgetNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Presupuesto no encontrado"}})
	case errors.Is(err, ErrInvalidPeriod):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_PERIOD", "message": "Período inválido (monthly|yearly)"}})
	case errors.Is(err, ErrInvalidAmount):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_AMOUNT", "message": "Monto debe ser mayor a cero"}})
	case errors.Is(err, ErrInvalidDate):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": "Fecha inválida (YYYY-MM-DD)"}})
	case errors.Is(err, ErrEndBeforeStart):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "END_BEFORE_START", "message": "end_date debe ser >= start_date"}})
	case errors.Is(err, ErrCategoryMissing):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "CATEGORY_NOT_FOUND", "message": "Categoría no encontrada o no pertenece al usuario"}})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
	}
}