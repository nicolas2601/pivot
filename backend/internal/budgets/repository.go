package budgets

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrBudgetNotFound = errors.New("budget not found")
)

type Repository interface {
	Create(b *Budget) error
	GetByID(id, userID uuid.UUID) (*Budget, error)
	ListByUser(userID uuid.UUID) ([]Budget, error)
	Update(b *Budget) error
	Delete(id, userID uuid.UUID) error
	FindByCategoryAndPeriod(userID, categoryID uuid.UUID, period Period) (*Budget, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) Create(b *Budget) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return r.db.Create(b).Error
}

func (r *repo) GetByID(id, userID uuid.UUID) (*Budget, error) {
	var b Budget
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBudgetNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *repo) ListByUser(userID uuid.UUID) ([]Budget, error) {
	var list []Budget
	err := r.db.Where("user_id = ?", userID).
		Order("start_date DESC, created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *repo) Update(b *Budget) error {
	return r.db.Save(b).Error
}

func (r *repo) Delete(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Budget{}).Error
}

func (r *repo) FindByCategoryAndPeriod(userID, categoryID uuid.UUID, period Period) (*Budget, error) {
	var b Budget
	err := r.db.Where("user_id = ? AND category_id = ? AND period = ?",
		userID, categoryID, period).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrBudgetNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}