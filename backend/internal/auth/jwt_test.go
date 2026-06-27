package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicolas/finanzas/backend/internal/auth"
)

func TestGenerateAndParseAccessToken_RoundTrip(t *testing.T) {
	secret := "test-secret-with-enough-entropy-32-chars-min"
	userID := uuid.New()
	email := "user@example.com"

	token, err := auth.GenerateAccessToken(userID, email, secret)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	claims, err := auth.ParseAccessToken(token, secret)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.WithinDuration(t, time.Now().Add(auth.AccessTokenDuration), claims.ExpiresAt.Time, 2*time.Second)
}

func TestGenerateAccessToken_EmptySecret_ReturnsError(t *testing.T) {
	_, err := auth.GenerateAccessToken(uuid.New(), "user@example.com", "")
	assert.Error(t, err)
}

func TestParseAccessToken_WrongSecret_ReturnsErrInvalidToken(t *testing.T) {
	token, err := auth.GenerateAccessToken(uuid.New(), "user@example.com", "secret-1-with-enough-entropy-32")
	require.NoError(t, err)

	_, err = auth.ParseAccessToken(token, "secret-2-different-with-enough-entropy")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestParseAccessToken_TamperedToken_ReturnsErrInvalidToken(t *testing.T) {
	token, err := auth.GenerateAccessToken(uuid.New(), "user@example.com", "secret-1234567890-entropy-padding")
	require.NoError(t, err)

	tampered := token + "tampered"

	_, err = auth.ParseAccessToken(tampered, "secret-1234567890-entropy-padding")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}

func TestParseAccessToken_ExpiredToken_ReturnsErrInvalidToken(t *testing.T) {
	secret := "secret-with-enough-entropy-32-chars-min-pad"

	claims := auth.Claims{
		UserID: uuid.New(),
		Email:  "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = auth.ParseAccessToken(signed, secret)
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}