package accounts

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	uid, _ := uuidParse(userID)
	a, err := h.svc.Create(uid, req)
	if err != nil {
		if errors.Is(err, ErrInvalidType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_TYPE", "message": "Tipo de cuenta inválido"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	uid, _ := uuidParse(userID)
	list, err := h.svc.List(uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, ListResponse{Accounts: list})
}

func (h *Handler) Get(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	uid, _ := uuidParse(userID)
	id, _ := uuidParse(c.Param("id"))
	a, err := h.svc.Get(id, uid)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Cuenta no encontrada"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *Handler) Update(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	uid, _ := uuidParse(userID)
	id, _ := uuidParse(c.Param("id"))
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	a, err := h.svc.Update(id, uid, req)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Cuenta no encontrada"}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.MustGet("userID").(string)
	uid, _ := uuidParse(userID)
	id, _ := uuidParse(c.Param("id"))
	if err := h.svc.Delete(id, uid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.Status(http.StatusNoContent)
}