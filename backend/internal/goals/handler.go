package goals

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const userIDContextKey = "userID"

type Handler struct{ svc *Service }

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// RegisterRoutes binds CRUD + deposit/withdraw endpoints. All require auth.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, requireUserID gin.HandlerFunc) {
	g := rg.Group("/goals", requireUserID)
	g.GET("", h.list)
	g.POST("", h.create)
	g.GET(":id", h.get)
	g.PATCH(":id", h.update)
	g.DELETE(":id", h.delete)
	g.POST(":id/deposit", h.deposit)
	g.POST(":id/withdraw", h.withdraw)
}

func (h *Handler) userID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get(userIDContextKey)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "MISSING_AUTH_CONTEXT", "message": "userID missing from request context"},
		})
		return uuid.Nil, false
	}
	s, ok := v.(string)
	if !ok || s == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INVALID_AUTH_CONTEXT", "message": "userID has wrong type"},
		})
		return uuid.Nil, false
	}
	id, err := uuid.Parse(s)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{"code": "BAD_USER_ID", "message": "userID is not a valid UUID"},
		})
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) list(c *gin.Context) {
	uid, ok := h.userID(c)
	if !ok {
		return
	}
	list, err := h.svc.List(uid)
	if err != nil {
		serverError(c, err)
		return
	}
	out := make([]*GoalDTO, 0, len(list))
	for _, g := range list {
		out = append(out, g.ToDTO())
	}
	c.JSON(http.StatusOK, gin.H{"goals": out})
}

func (h *Handler) create(c *gin.Context) {
	uid, ok := h.userID(c)
	if !ok {
		return
	}
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "VALIDATION_ERROR", "JSON inválido: "+err.Error())
		return
	}
	g, err := h.svc.Create(uid, req)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, g.ToDTO())
}

func (h *Handler) get(c *gin.Context) {
	uid, ok := h.userID(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	g, err := h.svc.Get(id, uid)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, g.ToDTO())
}

func (h *Handler) update(c *gin.Context) {
	uid, ok := h.userID(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "VALIDATION_ERROR", "JSON inválido: "+err.Error())
		return
	}
	g, err := h.svc.Update(id, uid, req)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, g.ToDTO())
}

func (h *Handler) delete(c *gin.Context) {
	uid, ok := h.userID(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(id, uid); err != nil {
		mapServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) deposit(c *gin.Context) {
	uid, ok := h.userID(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "VALIDATION_ERROR", "JSON inválido: "+err.Error())
		return
	}
	g, err := h.svc.Deposit(id, uid, req)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, g.ToDTO())
}

func (h *Handler) withdraw(c *gin.Context) {
	uid, ok := h.userID(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "id")
	if !ok {
		return
	}
	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "VALIDATION_ERROR", "JSON inválido: "+err.Error())
		return
	}
	g, err := h.svc.Withdraw(id, uid, req)
	if err != nil {
		mapServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, g.ToDTO())
}

// --- error mapping helpers (kept consistent with other handlers) ---

func parseID(c *gin.Context, key string) (uuid.UUID, bool) {
	raw := c.Param(key)
	id, err := uuid.Parse(raw)
	if err != nil {
		badRequest(c, "INVALID_ID", "id must be a UUID")
		return uuid.Nil, false
	}
	return id, true
}

func badRequest(c *gin.Context, code, msg string) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error": gin.H{"code": code, "message": msg},
	})
}

func serverError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
		"error": gin.H{"code": "INTERNAL_ERROR", "message": err.Error()},
	})
}

func mapServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrGoalNotFound):
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "goal not found"},
		})
	case errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrInvalidName),
		errors.Is(err, ErrInvalidCurrency),
		errors.Is(err, ErrInvalidDeadline),
		errors.Is(err, ErrNonPositiveMove),
		errors.Is(err, ErrOverWithdraw),
		errors.Is(err, ErrGoalAlreadyDone):
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
	default:
		serverError(c, err)
	}
}
