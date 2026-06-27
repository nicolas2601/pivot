package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/testhelpers"
)

func setupHandler(t *testing.T) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)
	t.Helper()

	db := testhelpers.SetupTestDB(t)
	repo := auth.NewUserRepository(db.DB)
	sessions := auth.NewSessionRepository(db.DB)
	cfg := &config.Config{JWTSecret: "test-secret-with-enough-entropy-32-chars-min", GinMode: "test"}
	svc := auth.NewService(repo, sessions, cfg)

	r := gin.New()
	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, svc, cfg)

	gin.DefaultWriter = &bytes.Buffer{}

	return r, db.Cleanup
}

func TestHandler_Register_Success(t *testing.T) {
	r, cleanup := setupHandler(t)
	defer cleanup()

	body := `{"email":"user@example.com","password":"password123","display_name":"Test"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp auth.AuthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "user@example.com", resp.User.Email)
	assert.NotEmpty(t, resp.AccessToken)
	assert.NotEmpty(t, resp.RefreshToken)

	cookies := w.Result().Cookies()
	var refreshCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshCookie = cookie
		}
	}
	require.NotNil(t, refreshCookie)
	assert.True(t, refreshCookie.HttpOnly)
	assert.NotEmpty(t, refreshCookie.Value)
}

func TestHandler_Login_WrongPassword_Returns401(t *testing.T) {
	r, cleanup := setupHandler(t)
	defer cleanup()

	regBody := `{"email":"user@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	loginBody := `{"email":"user@example.com","password":"wrongpassword"}`
	req = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Register_DuplicateEmail_Returns409(t *testing.T) {
	r, cleanup := setupHandler(t)
	defer cleanup()

	body := `{"email":"dup@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	req = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestHandler_Me_RequiresAuth(t *testing.T) {
	r, cleanup := setupHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Me_WithValidToken_ReturnsUser(t *testing.T) {
	r, cleanup := setupHandler(t)
	defer cleanup()

	body := `{"email":"user@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var regResp auth.AuthResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &regResp))

	req = httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+regResp.AccessToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var meResp auth.UserResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &meResp))
	assert.Equal(t, "user@example.com", meResp.User.Email)
}

func TestHandler_Logout_ClearsCookie(t *testing.T) {
	r, cleanup := setupHandler(t)
	defer cleanup()

	body := `{"email":"user@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	req = httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	cookies := w.Result().Cookies()
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			assert.Equal(t, "", cookie.Value)
			assert.Equal(t, -1, cookie.MaxAge)
		}
	}
}