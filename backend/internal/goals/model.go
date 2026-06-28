package goals

import (
	"time"

	"github.com/google/uuid"
)

// Goal represents a savings goal owned by a user.
type Goal struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Name          string     `gorm:"not null;size:100" json:"name"`
	TargetAmount  int64      `gorm:"not null" json:"target_amount"`
	CurrentAmount int64      `gorm:"not null;default:0" json:"current_amount"`
	Currency      string     `gorm:"not null;size:3;default:COP" json:"currency"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	AccountID     *uuid.UUID `gorm:"type:uuid;index" json:"account_id,omitempty"`
	Color         *string    `gorm:"size:7" json:"color,omitempty"`
	Notes         *string    `gorm:"type:text" json:"notes,omitempty"`
	IsCompleted   bool       `gorm:"not null;default:false" json:"is_completed"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// PercentComplete returns 0-100 (clamped) so the UI can render a progress bar
// without re-doing the math. Returns 100 once the goal is at or above target.
func (g *Goal) PercentComplete() int {
	if g.TargetAmount <= 0 {
		return 0
	}
	pct := int((g.CurrentAmount * 100) / g.TargetAmount)
	if pct < 0 {
		return 0
	}
	if pct > 100 {
		return 100
	}
	return pct
}

func (Goal) TableName() string { return "goals" }
