package auth_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/testhelpers"
)

func setupService(t *testing.T) (auth.Service, auth.UserRepository, auth.SessionRepository, func()) {
	t.Helper()
	db := testhelpers.SetupTestDB(t)

	repo := auth.NewUserRepository(db.DB)
	sessions := auth.NewSessionRepository(db.DB)
	cfg := &config.Config{JWTSecret: "test-secret-with-enough-entropy-32-chars-min"}
	svc := auth.NewService(repo, sessions, cfg)

	return svc, repo, sessions, db.Cleanup
}

func TestService_Register_Success(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	user, err := svc.Register("user@example.com", "password123", "Test User")
	require.NoError(t, err)
	assert.NotNil(t, user.ID)
	assert.Equal(t, "user@example.com", user.Email)
	assert.NotNil(t, user.DisplayName)
	assert.Equal(t, "Test User", *user.DisplayName)
	assert.NotEmpty(t, user.PasswordHash) // bcrypt hash is non-empty

	// Verify PasswordHash is excluded from JSON serialization (security)
	jsonBytes, err := json.Marshal(user)
	require.NoError(t, err)
	assert.NotContains(t, string(jsonBytes), user.PasswordHash, "password hash must not be exposed in JSON")
}

func TestService_Register_DuplicateEmail_ReturnsError(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, err := svc.Register("dup@example.com", "password123", "")
	require.NoError(t, err)

	_, err = svc.Register("dup@example.com", "password123", "")
	assert.ErrorIs(t, err, auth.ErrUserAlreadyExists)
}

func TestService_Register_ShortPassword_ReturnsErrInvalidInput(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, err := svc.Register("user@example.com", "short", "")
	assert.ErrorIs(t, err, auth.ErrInvalidInput)
}

func TestService_Register_EmptyEmail_ReturnsErrInvalidInput(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, err := svc.Register("", "password123", "")
	assert.ErrorIs(t, err, auth.ErrInvalidInput)
}

func TestService_Login_Success(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, err := svc.Register("user@example.com", "password123", "")
	require.NoError(t, err)

	result, err := svc.Login("user@example.com", "password123", "test-agent", "127.0.0.1")
	require.NoError(t, err)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, "user@example.com", result.User.Email)
}

func TestService_Login_WrongPassword_ReturnsErrInvalidCredentials(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, _ = svc.Register("user@example.com", "password123", "")

	_, err := svc.Login("user@example.com", "wrongpassword", "", "")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestService_Login_NonexistentUser_ReturnsErrInvalidCredentials(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, err := svc.Login("nobody@example.com", "password123", "", "")
	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestService_Refresh_RotatesRefreshToken(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, _ = svc.Register("user@example.com", "password123", "")
	login, err := svc.Login("user@example.com", "password123", "", "")
	require.NoError(t, err)

	refreshed, err := svc.Refresh(login.RefreshToken)
	require.NoError(t, err)
	assert.NotEqual(t, login.RefreshToken, refreshed.RefreshToken)
	assert.Equal(t, "user@example.com", refreshed.User.Email)

	// Old refresh token should now be revoked
	_, err = svc.Refresh(login.RefreshToken)
	assert.ErrorIs(t, err, auth.ErrSessionRevoked)
}

func TestService_Refresh_InvalidToken_ReturnsErrInvalidRefreshToken(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, err := svc.Refresh("not-a-real-refresh-token")
	assert.ErrorIs(t, err, auth.ErrInvalidRefreshToken)
}

func TestService_Logout_RevokesSession(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, _ = svc.Register("user@example.com", "password123", "")
	login, err := svc.Login("user@example.com", "password123", "", "")
	require.NoError(t, err)

	err = svc.Logout(login.RefreshToken)
	require.NoError(t, err)

	_, err = svc.Refresh(login.RefreshToken)
	assert.ErrorIs(t, err, auth.ErrSessionRevoked)
}

func TestService_Logout_UnknownToken_IsIdempotent(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	err := svc.Logout("never-existed-token")
	assert.NoError(t, err)
}

func TestService_Me_Success(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	registered, _ := svc.Register("user@example.com", "password123", "")
	login, err := svc.Login("user@example.com", "password123", "", "")
	require.NoError(t, err)

	me, err := svc.Me(login.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, registered.ID, me.ID)
}

func TestService_Me_InvalidToken_ReturnsErrInvalidToken(t *testing.T) {
	svc, _, _, cleanup := setupService(t)
	defer cleanup()

	_, err := svc.Me("not-a-valid-jwt")
	assert.ErrorIs(t, err, auth.ErrInvalidToken)
}