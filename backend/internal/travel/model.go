package travel

import (
	"time"

	"github.com/google/uuid"
)

type MemberRole string

const (
	RoleOwner  MemberRole = "owner"
	RoleMember MemberRole = "member"
)

type SettlementStatus string

const (
	SettlementPending   SettlementStatus = "pending"
	SettlementConfirmed SettlementStatus = "confirmed"
)

type SplitMethod string

const (
	SplitEqual       SplitMethod = "equal"
	SplitExact       SplitMethod = "exact"
	SplitPercentage  SplitMethod = "percentage"
)

type TravelGroup struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string    `gorm:"not null;size:100" json:"name"`
	Description *string   `json:"description,omitempty"`
	Currency    string    `gorm:"not null;size:3;default:COP" json:"currency"`
	CreatedBy   uuid.UUID `gorm:"type:uuid;not null;index" json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (TravelGroup) TableName() string {
	return "travel_groups"

}

type TravelGroupMember struct {
	ID       uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	GroupID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"group_id"`
	UserID   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Role     MemberRole `gorm:"not null;size:20;default:member" json:"role"`
	JoinedAt time.Time  `json:"joined_at"`
}

func (TravelGroupMember) TableName() string {
	return "travel_group_members"
}

type TravelExpense struct {
	ID          uuid.UUID   `gorm:"type:uuid;primaryKey" json:"id"`
	GroupID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"group_id"`
	PaidBy      uuid.UUID   `gorm:"type:uuid;not null;index" json:"paid_by"`
	Amount      int64       `gorm:"not null" json:"amount"`
	Currency    string      `gorm:"not null;size:3;default:COP" json:"currency"`
	Description string      `gorm:"not null;size:255" json:"description"`
	SplitMethod SplitMethod `gorm:"not null;size:20" json:"split_method"`
	Date        time.Time   `gorm:"not null" json:"date"`
	CreatedAt   time.Time   `json:"created_at"`
}

func (TravelExpense) TableName() string {
	return "travel_expenses"
}

type TravelExpenseShare struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ExpenseID uuid.UUID `gorm:"type:uuid;not null;index" json:"expense_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Amount    int64     `gorm:"not null" json:"amount"`
}

func (TravelExpenseShare) TableName() string {
	return "travel_expense_shares"
}

type TravelSettlement struct {
	ID          uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	GroupID     uuid.UUID        `gorm:"type:uuid;not null;index" json:"group_id"`
	FromUser    uuid.UUID        `gorm:"type:uuid;not null;index" json:"from_user"`
	ToUser      uuid.UUID        `gorm:"type:uuid;not null;index" json:"to_user"`
	Amount      int64            `gorm:"not null" json:"amount"`
	Currency    string           `gorm:"not null;size:3;default:COP" json:"currency"`
	Status      SettlementStatus `gorm:"not null;size:20;default:pending" json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	ConfirmedAt *time.Time       `json:"confirmed_at,omitempty"`
}

func (TravelSettlement) TableName() string {
	return "travel_settlements"
}

func IsValidRole(r string) bool {
	switch MemberRole(r) {
	case RoleOwner, RoleMember:
		return true
	}
	return false
}

func IsValidSplitMethod(m string) bool {
	switch SplitMethod(m) {
	case SplitEqual, SplitExact, SplitPercentage:
		return true
	}
	return false
}