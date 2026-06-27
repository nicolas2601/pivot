package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/nicolas/finanzas/backend/internal/config"
)

const (
	RefreshTokenDuration = 30 * 24 * time.Hour
	bcryptCost           = 12
)

type LoginResult struct {
	User         *User
	AccessToken  string
	RefreshToken string
}

type Service interface {
	Register(email, password, displayName string) (*User, error)
	Login(email, password, userAgent, ip string) (*LoginResult, error)
	Refresh(refreshToken string) (*LoginResult, error)
	Logout(refreshToken string) error
	Me(accessToken string) (*User, error)
}

type service struct {
	repo      UserRepository
	sessions  SessionRepository
	jwtSecret string
}

func NewService(repo UserRepository, sessions SessionRepository, cfg *config.Config) Service {
	return &service{
		repo:      repo,
		sessions:  sessions,
		jwtSecret: cfg.JWTSecret,
	}
}

func (s *service) Register(email, password, displayName string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || len(password) < 8 {
		return nil, ErrInvalidInput
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &User{
		Email:        email,
		PasswordHash: string(hash),
	}
	if displayName != "" {
		name := displayName
		user.DisplayName = &name
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *service) Login(email, password, userAgent, ip string) (*LoginResult, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := GenerateAccessToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	session := &Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		UserAgent:        userAgent,
		IPAddress:        ip,
		ExpiresAt:        time.Now().Add(RefreshTokenDuration),
	}
	if err := s.sessions.Create(session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *service) Refresh(refreshToken string) (*LoginResult, error) {
	hash := hashRefreshToken(refreshToken)
	session, err := s.sessions.FindByRefreshTokenHash(hash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if session.RevokedAt != nil {
		return nil, ErrSessionRevoked
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	user, err := s.repo.FindByID(session.UserID)
	if err != nil {
		return nil, err
	}

	// Rotate: revoke old, issue new
	if err := s.sessions.Revoke(session.ID); err != nil {
		return nil, fmt.Errorf("revoke old session: %w", err)
	}

	accessToken, err := GenerateAccessToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	newSession := &Session{
		ID:               uuid.New(),
		UserID:           user.ID,
		RefreshTokenHash: hashRefreshToken(newRefreshToken),
		ExpiresAt:        time.Now().Add(RefreshTokenDuration),
	}
	if err := s.sessions.Create(newSession); err != nil {
		return nil, fmt.Errorf("create new session: %w", err)
	}

	return &LoginResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *service) Logout(refreshToken string) error {
	hash := hashRefreshToken(refreshToken)
	session, err := s.sessions.FindByRefreshTokenHash(hash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil
		}
		return err
	}
	return s.sessions.Revoke(session.ID)
}

func (s *service) Me(accessToken string) (*User, error) {
	claims, err := ParseAccessToken(accessToken, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByID(claims.UserID)
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(h[:])
}