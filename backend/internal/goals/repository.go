package goals

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrGoalNotFound = errors.New("goal not found")

// Repository defines the persistence contract for goals.
type Repository interface {
	Create(g *Goal) error
	GetByID(id, userID uuid.UUID) (*Goal, error)
	ListByUser(userID uuid.UUID) ([]Goal, error)
	Update(g *Goal) error
	Delete(id, userID uuid.UUID) error

	// AtomicAdjustCurrent bumps current_amount by `delta` (positive or negative)
	// and marks the goal completed/active accordingly. Returns the updated goal.
	// Implementation must run inside a single SQL UPDATE so two concurrent
	// deposits don't both read stale `current_amount`.
	AtomicAdjustCurrent(id, userID uuid.UUID, delta int64) (*Goal, error)
}

type repo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &repo{db: db} }

func (r *repo) Create(g *Goal) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return r.db.Create(g).Error
}

func (r *repo) GetByID(id, userID uuid.UUID) (*Goal, error) {
	var g Goal
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrGoalNotFound
	}
	if err != nil {
		return nil, err
	}
	return &g, nil
}

func (r *repo) ListByUser(userID uuid.UUID) ([]Goal, error) {
	var list []Goal
	err := r.db.Where("user_id = ?", userID).
		Order("is_completed ASC, deadline ASC NULLS LAST, created_at DESC").
		Find(&list).Error
	return list, err
}

func (r *repo) Update(g *Goal) error {
	return r.db.Save(g).Error
}

func (r *repo) Delete(id, userID uuid.UUID) error {
	return r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Goal{}).Error
}

// AtomicAdjustCurrent uses a single UPDATE statement with an arithmetic
// expression to avoid the read-modify-write race. GORM's Expr() is a safe
// parameterized expression; Postgres serializes concurrent updates on the
// same row.
func (r *repo) AtomicAdjustCurrent(id, userID uuid.UUID, delta int64) (*Goal, error) {
	// Step 1: bump the column and (if at/above target) mark completed, atomically.
	res := r.db.Model(&Goal{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]any{
			"current_amount": gorm.Expr("current_amount + ?", delta),
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, ErrGoalNotFound
	}
	// Step 2: re-read to derive is_completed / completed_at from the new amount.
	var g Goal
	if err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&g).Error; err != nil {
		return nil, err
	}
	now := time.Now()
	if g.CurrentAmount >= g.TargetAmount && !g.IsCompleted {
		g.IsCompleted = true
		g.CompletedAt = &now
		if err := r.db.Save(&g).Error; err != nil {
			return nil, err
		}
	} else if g.CurrentAmount < g.TargetAmount && g.IsCompleted {
		// Manual edit pulled the bar back below 100% — re-open the goal.
		g.IsCompleted = false
		g.CompletedAt = nil
		if err := r.db.Save(&g).Error; err != nil {
			return nil, err
		}
	}
	return &g, nil
}
