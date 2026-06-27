package accounts

import (
	"time"

	"github.com/google/uuid"
)

type AccountType string

const (
	TypeCash    AccountType = "cash"
	TypeDebit   AccountType = "debit"
	TypeCredit  AccountType = "credit"
	TypeSavings AccountType = "savings"
)

type Account struct {
	ID             uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID   `gorm:"type:uuid;not null;index" json:"user_id"`
	Name           string      `gorm:"not null;size:100" json:"name"`
	Type           AccountType `gorm:"not null;size:20" json:"type"`
	Currency       string      `gorm:"not null;size:3;default:COP" json:"currency"`
	OpeningBalance int64       `gorm:"not null;default:0" json:"opening_balance"`
	Color          *string     `gorm:"size:7" json:"color,omitempty"`
	Icon           *string     `gorm:"size:50" json:"icon,omitempty"`
	DeletedAt      *time.Time  `gorm:"index" json:"-"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

func (Account) TableName() string {
	return "accounts"
}