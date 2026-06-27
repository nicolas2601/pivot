package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrSessionExpired      = errors.New("session expired")
	ErrSessionRevoked       = errors.New("session revoked")
	ErrInvalidInput         = errors.New("invalid input")
	ErrSessionNotFound      = errors.New("session not found")
)