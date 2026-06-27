package travel

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrGroupNotFound     = errors.New("travel group not found")
	ErrExpenseNotFound   = errors.New("travel expense not found")
	ErrMemberNotFound    = errors.New("member not found in group")
	ErrAlreadyMember     = errors.New("user is already a member of this group")
	ErrNotMember         = errors.New("user is not a member of this group")
	ErrInvalidSplit      = errors.New("invalid split method")
	ErrSplitSumMismatch  = errors.New("split amounts must sum to expense total")
	ErrSplitUsersMissing = errors.New("at least one user must be included in the split")
	ErrPayerNotMember    = errors.New("payer must be a member of the group")
	ErrCannotRemoveOwner = errors.New("cannot remove the group owner")
)

type Repository interface {
	// Groups
	CreateGroup(g *TravelGroup) error
	GetGroup(id uuid.UUID) (*TravelGroup, error)
	ListGroupsByUser(userID uuid.UUID) ([]TravelGroup, error)
	UpdateGroup(g *TravelGroup) error
	DeleteGroup(id uuid.UUID) error

	// Members
	AddMember(m *TravelGroupMember) error
	ListMembers(groupID uuid.UUID) ([]TravelGroupMember, error)
	GetMember(groupID, userID uuid.UUID) (*TravelGroupMember, error)
	RemoveMember(groupID, userID uuid.UUID) error
	CountOwners(groupID uuid.UUID) (int, error)

	// Expenses
	CreateExpenseWithShares(expense *TravelExpense, shares []TravelExpenseShare) error
	GetExpense(id uuid.UUID) (*TravelExpense, error)
	ListExpensesByGroup(groupID uuid.UUID) ([]TravelExpense, error)
	DeleteExpense(id uuid.UUID) error
	ListSharesByExpense(expenseID uuid.UUID) ([]TravelExpenseShare, error)

	// Settlements
	CreateSettlement(s *TravelSettlement) error
	GetSettlement(id uuid.UUID) (*TravelSettlement, error)
	ListSettlementsByGroup(groupID uuid.UUID) ([]TravelSettlement, error)
	UpdateSettlement(s *TravelSettlement) error

	// Aggregations
	SumPaidByUser(groupID, userID uuid.UUID) (int64, error)
	SumShareByUser(groupID, userID uuid.UUID) (int64, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

// --- Groups ---

func (r *repo) CreateGroup(g *TravelGroup) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return r.db.Create(g).Error
}

func (r *repo) GetGroup(id uuid.UUID) (*TravelGroup, error) {
	var g TravelGroup
	err := r.db.First(&g, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGroupNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *repo) ListGroupsByUser(userID uuid.UUID) ([]TravelGroup, error) {
	var groups []TravelGroup
	err := r.db.
		Joins("JOIN travel_group_members m ON m.group_id = travel_groups.id").
		Where("m.user_id = ?", userID).
		Order("travel_groups.created_at DESC").
		Find(&groups).Error
	return groups, err
}

func (r *repo) UpdateGroup(g *TravelGroup) error {
	return r.db.Save(g).Error
}

func (r *repo) DeleteGroup(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&TravelGroup{}).Error
}

// --- Members ---

func (r *repo) AddMember(m *TravelGroupMember) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.JoinedAt.IsZero() {
		m.JoinedAt = time.Now()
	}
	return r.db.Create(m).Error
}

func (r *repo) ListMembers(groupID uuid.UUID) ([]TravelGroupMember, error) {
	var ms []TravelGroupMember
	err := r.db.Where("group_id = ?", groupID).
		Order("joined_at ASC").
		Find(&ms).Error
	return ms, err
}

func (r *repo) GetMember(groupID, userID uuid.UUID) (*TravelGroupMember, error) {
	var m TravelGroupMember
	err := r.db.Where("group_id = ? AND user_id = ?", groupID, userID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMemberNotFound
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *repo) RemoveMember(groupID, userID uuid.UUID) error {
	return r.db.Where("group_id = ? AND user_id = ?", groupID, userID).
		Delete(&TravelGroupMember{}).Error
}

func (r *repo) CountOwners(groupID uuid.UUID) (int, error) {
	var count int64
	err := r.db.Model(&TravelGroupMember{}).
		Where("group_id = ? AND role = ?", groupID, RoleOwner).
		Count(&count).Error
	return int(count), err
}

// --- Expenses ---

func (r *repo) CreateExpenseWithShares(expense *TravelExpense, shares []TravelExpenseShare) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if expense.ID == uuid.Nil {
			expense.ID = uuid.New()
		}
		if err := tx.Create(expense).Error; err != nil {
			return err
		}
		for i := range shares {
			if shares[i].ID == uuid.Nil {
				shares[i].ID = uuid.New()
			}
			shares[i].ExpenseID = expense.ID
			if err := tx.Create(&shares[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *repo) GetExpense(id uuid.UUID) (*TravelExpense, error) {
	var e TravelExpense
	err := r.db.First(&e, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrExpenseNotFound
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *repo) ListExpensesByGroup(groupID uuid.UUID) ([]TravelExpense, error) {
	var es []TravelExpense
	err := r.db.Where("group_id = ?", groupID).
		Order("date DESC, created_at DESC").
		Find(&es).Error
	return es, err
}

func (r *repo) DeleteExpense(id uuid.UUID) error {
	return r.db.Where("id = ?", id).Delete(&TravelExpense{}).Error
}

func (r *repo) ListSharesByExpense(expenseID uuid.UUID) ([]TravelExpenseShare, error) {
	var ss []TravelExpenseShare
	err := r.db.Where("expense_id = ?", expenseID).Find(&ss).Error
	return ss, err
}

// --- Settlements ---

func (r *repo) CreateSettlement(s *TravelSettlement) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = SettlementPending
	}
	return r.db.Create(s).Error
}

func (r *repo) GetSettlement(id uuid.UUID) (*TravelSettlement, error) {
	var s TravelSettlement
	err := r.db.First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrExpenseNotFound
	}
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *repo) ListSettlementsByGroup(groupID uuid.UUID) ([]TravelSettlement, error) {
	var ss []TravelSettlement
	err := r.db.Where("group_id = ?", groupID).
		Order("created_at DESC").
		Find(&ss).Error
	return ss, err
}

func (r *repo) UpdateSettlement(s *TravelSettlement) error {
	return r.db.Save(s).Error
}

// --- Aggregations ---

func (r *repo) SumPaidByUser(groupID, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.Model(&TravelExpense{}).
		Where("group_id = ? AND paid_by = ?", groupID, userID).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *repo) SumShareByUser(groupID, userID uuid.UUID) (int64, error) {
	var total int64
	err := r.db.Model(&TravelExpenseShare{}).
		Joins("JOIN travel_expenses e ON e.id = travel_expense_shares.expense_id").
		Where("e.group_id = ? AND travel_expense_shares.user_id = ?", groupID, userID).
		Select("COALESCE(SUM(travel_expense_shares.amount), 0)").
		Scan(&total).Error
	return total, err
}