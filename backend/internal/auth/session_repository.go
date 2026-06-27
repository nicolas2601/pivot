package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
}

type SessionRepository interface {
	Create(session *Session) error
	FindByRefreshTokenHash(hash string) (*Session, error)
	Revoke(id uuid.UUID) error
}

type sessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) SessionRepository {
	return &sessionRepository{db: db}
}

type sessionModel struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index"`
	RefreshTokenHash string    `gorm:"not null;size:255;index"`
	UserAgent        string
	IPAddress        string    `gorm:"size:45"`
	ExpiresAt        time.Time `gorm:"not null;index"`
	RevokedAt        *time.Time
	CreatedAt        time.Time
}

func (sessionModel) TableName() string {
	return "sessions"
}

func (r *sessionRepository) Create(session *Session) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	model := sessionModel{
		ID:               session.ID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		UserAgent:        session.UserAgent,
		IPAddress:        session.IPAddress,
		ExpiresAt:        session.ExpiresAt,
		CreatedAt:        time.Now(),
	}
	return r.db.Create(&model).Error
}

func (r *sessionRepository) FindByRefreshTokenHash(hash string) (*Session, error) {
	var model sessionModel
	err := r.db.Where("refresh_token_hash = ?", hash).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &Session{
		ID:               model.ID,
		UserID:           model.UserID,
		RefreshTokenHash: model.RefreshTokenHash,
		UserAgent:        model.UserAgent,
		IPAddress:        model.IPAddress,
		ExpiresAt:        model.ExpiresAt,
		RevokedAt:        model.RevokedAt,
	}, nil
}

func (r *sessionRepository) Revoke(id uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&sessionModel{}).Where("id = ?", id).Update("revoked_at", &now).Error
}