package accounts

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository interface {
	Create(a *Account) error
	ListByUser(userID uuid.UUID) ([]Account, error)
	GetByID(id, userID uuid.UUID) (*Account, error)
	Update(a *Account) error
	Delete(id, userID uuid.UUID) error
}

type accountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &accountRepository{db: db}
}

func (r *accountRepository) Create(a *Account) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return r.db.Create(a).Error
}

func (r *accountRepository) ListByUser(userID uuid.UUID) ([]Account, error) {
	var list []Account
	err := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at ASC").
		Find(&list).Error
	return list, err
}

func (r *accountRepository) GetByID(id, userID uuid.UUID) (*Account, error) {
	var a Account
	err := r.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *accountRepository) Update(a *Account) error {
	return r.db.Save(a).Error
}

func (r *accountRepository) Delete(id, userID uuid.UUID) error {
	return r.db.Model(&Account{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}