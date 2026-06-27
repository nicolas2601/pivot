package budgets

import (
	"time"

	"github.com/google/uuid"
)

type Period string

const (
	PeriodMonthly Period = "monthly"
	PeriodYearly  Period = "yearly"
)

type Budget struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	CategoryID uuid.UUID  `gorm:"type:uuid;not null;index" json:"category_id"`
	Amount     int64      `gorm:"not null" json:"amount"`
	Period     Period     `gorm:"not null;size:20;default:monthly" json:"period"`
	StartDate  time.Time  `gorm:"not null" json:"start_date"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

func (Budget) TableName() string {
	return "budgets"
}

func IsValidPeriod(p string) bool {
	switch Period(p) {
	case PeriodMonthly, PeriodYearly:
		return true
	}
	return false
}