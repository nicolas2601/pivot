package travel

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

// --- Groups ---

func (h *Handler) CreateGroup(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	var dto CreateGroupDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	g, err := h.svc.CreateGroup(userID, dto.ToServiceCreate())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, GroupResponse{Group: *g})
}

func (h *Handler) ListGroups(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	gs, err := h.svc.ListGroups(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
		return
	}
	c.JSON(http.StatusOK, GroupsResponse{Groups: gs})
}

func (h *Handler) GetGroup(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	id, _ := uuid.Parse(c.Param("id"))
	g, err := h.svc.GetGroup(id, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, GroupResponse{Group: *g})
}

func (h *Handler) UpdateGroup(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	id, _ := uuid.Parse(c.Param("id"))
	var dto UpdateGroupDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	g, err := h.svc.UpdateGroup(id, userID, dto.ToServiceUpdate())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, GroupResponse{Group: *g})
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	id, _ := uuid.Parse(c.Param("id"))
	if err := h.svc.DeleteGroup(id, userID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Members ---

func (h *Handler) AddMember(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	var dto AddMemberDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	m, err := h.svc.AddMemberByEmail(groupID, userID, AddMemberRequest{
		Email: dto.Email,
		Role:  dto.Role,
	})
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *Handler) ListMembers(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	ms, err := h.svc.ListMembers(groupID, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, MembersResponse{Members: ms})
}

func (h *Handler) RemoveMember(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	target, _ := uuid.Parse(c.Param("user_id"))
	if err := h.svc.RemoveMember(groupID, target, userID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Expenses ---

func (h *Handler) AddExpense(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	var dto AddExpenseDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	req, err := dto.ToServiceAdd()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}
	exp, shares, err := h.svc.AddExpense(groupID, userID, req)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, ExpenseResponse{Expense: *exp, Shares: shares})
}

func (h *Handler) ListExpenses(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	es, err := h.svc.ListExpenses(groupID, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ExpensesResponse{Expenses: es})
}

func (h *Handler) GetExpense(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	expenseID, _ := uuid.Parse(c.Param("expense_id"))
	exp, shares, err := h.svc.GetExpense(expenseID, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, ExpenseResponse{Expense: *exp, Shares: shares})
}

func (h *Handler) DeleteExpense(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	expenseID, _ := uuid.Parse(c.Param("expense_id"))
	if err := h.svc.DeleteExpense(expenseID, userID); err != nil {
		writeServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Settlements ---

func (h *Handler) GetSettlements(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	ss, err := h.svc.ComputeSettlements(groupID, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, SettlementSuggestionResponse{Suggestions: ss})
}

func (h *Handler) RecordSettlement(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	var dto RecordSettlementDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()}})
		return
	}
	rec, err := h.svc.RecordSettlement(groupID, userID, dto.ToServiceRecord())
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, rec)
}

func (h *Handler) ListSettlements(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	groupID, _ := uuid.Parse(c.Param("id"))
	ss, err := h.svc.ListSettlements(groupID, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, SettlementsResponse{Settlements: ss})
}

func (h *Handler) ConfirmSettlement(c *gin.Context) {
	userID, _ := uuid.Parse(c.GetString("userID"))
	settlementID, _ := uuid.Parse(c.Param("settlement_id"))
	rec, err := h.svc.ConfirmSettlement(settlementID, userID)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, rec)
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrGroupNotFound), errors.Is(err, ErrExpenseNotFound), errors.Is(err, ErrMemberNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Recurso no encontrado"}})
	case errors.Is(err, ErrAlreadyMember):
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "ALREADY_MEMBER", "message": "El usuario ya es miembro"}})
	case errors.Is(err, ErrNotMember):
		c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "NOT_MEMBER", "message": "No eres miembro del grupo"}})
	case errors.Is(err, ErrInvalidRole):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_ROLE", "message": "Rol inválido"}})
	case errors.Is(err, ErrInvalidSplit):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_SPLIT", "message": "Método de división inválido"}})
	case errors.Is(err, ErrSplitSumMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "SPLIT_SUM_MISMATCH", "message": err.Error()}})
	case errors.Is(err, ErrSplitUsersMissing):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "NO_SPLIT_USERS", "message": "Debes incluir al menos un usuario en la división"}})
	case errors.Is(err, ErrPayerNotMember):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "PAYER_NOT_MEMBER", "message": "Quien paga debe ser miembro del grupo"}})
	case errors.Is(err, ErrCannotRemoveOwner):
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "CANNOT_REMOVE_OWNER", "message": "No puedes eliminar al único dueño"}})
	case errors.Is(err, ErrInvalidCurrency):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_CURRENCY", "message": "Código de moneda inválido (3 letras)"}})
	case errors.Is(err, ErrUserNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "USER_NOT_FOUND", "message": "Usuario no encontrado"}})
	case errors.Is(err, ErrInvalidDate):
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_DATE", "message": "Fecha inválida (YYYY-MM-DD)"}})
	default:
		msg := err.Error()
		// Service-level guard rails that we surface as 400s without their
		// own typed error: amount <= 0, same-user settlement, owner-only ops.
		if contains(msg, "amount must be greater than zero") ||
			contains(msg, "from_user and to_user must differ") ||
			contains(msg, "owner role required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": msg}})
			return
		}
		if contains(msg, "only the recipient can confirm") {
			c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "FORBIDDEN", "message": msg}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "INTERNAL", "message": "Error interno"}})
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}