package accounts_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicolas/finanzas/backend/internal/accounts"
	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/middleware"
	"github.com/nicolas/finanzas/backend/internal/testhelpers"
)

func setupAccountHandler(t *testing.T) (*gin.Engine, *auth.User, func()) {
	gin.SetMode(gin.TestMode)
	t.Helper()
	db := testhelpers.SetupTestDB(t)

	userRepo := auth.NewUserRepository(db.DB)
	email := uuid.NewString() + "@example.com"
	require.NoError(t, userRepo.Create(&auth.User{Email: email, PasswordHash: "x"}))
	user, err := userRepo.FindByEmail(email)
	require.NoError(t, err)

	accRepo := accounts.NewAccountRepository(db.DB)
	accSvc := accounts.NewService(accRepo)

	r := gin.New()
	api := r.Group("/api/v1")
	resolver := func(tok string) (string, error) {
		claims, err := auth.ParseAccessToken(tok, testSecret)
		if err != nil {
			return "", err
		}
		return claims.UserID.String(), nil
	}
	accounts.RegisterRoutes(api, accounts.NewHandler(accSvc), middleware.RequireUserID(resolver))

	return r, user, db.Cleanup
}

const testSecret = "test-secret-with-enough-entropy-32-chars-min"

func doRequest(r *gin.Engine, method, path, token string, body any) *httptest.ResponseRecorder {
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getToken(t *testing.T, r *gin.Engine, user *auth.User) string {
	token, err := auth.GenerateAccessToken(user.ID, user.Email, testSecret)
	require.NoError(t, err)
	return token
}

func TestHandler_Accounts_CRUD(t *testing.T) {
	r, user, cleanup := setupAccountHandler(t)
	defer cleanup()
	tok := getToken(t, r, user)

	// List empty
	w := doRequest(r, "GET", "/api/v1/accounts", tok, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	var listResp accounts.ListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	assert.Empty(t, listResp.Accounts)

	// Create
	createBody := accounts.CreateRequest{
		Name:     "Bancolombia",
		Type:     "debit",
		Currency: "COP",
	}
	w = doRequest(r, "POST", "/api/v1/accounts", tok, createBody)
	assert.Equal(t, http.StatusCreated, w.Code)
	var created accounts.Account
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.Equal(t, "Bancolombia", created.Name)

	// List now has 1
	w = doRequest(r, "GET", "/api/v1/accounts", tok, nil)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	assert.Len(t, listResp.Accounts, 1)

	// Get
	w = doRequest(r, "GET", "/api/v1/accounts/"+created.ID.String(), tok, nil)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update
	newName := "Bancolombia Updated"
	w = doRequest(r, "PATCH", "/api/v1/accounts/"+created.ID.String(), tok, map[string]string{"name": newName})
	assert.Equal(t, http.StatusOK, w.Code)
	var updated accounts.Account
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, newName, updated.Name)

	// Delete
	w = doRequest(r, "DELETE", "/api/v1/accounts/"+created.ID.String(), tok, nil)
	assert.Equal(t, http.StatusNoContent, w.Code)

	// Get after delete -> 404
	w = doRequest(r, "GET", "/api/v1/accounts/"+created.ID.String(), tok, nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_Accounts_RequiresAuth(t *testing.T) {
	r, _, cleanup := setupAccountHandler(t)
	defer cleanup()

	w := doRequest(r, "GET", "/api/v1/accounts", "", nil)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandler_Accounts_OwnershipEnforced(t *testing.T) {
	r, user, cleanup := setupAccountHandler(t)
	defer cleanup()
	tok := getToken(t, r, user)

	// Create as real user
	w := doRequest(r, "POST", "/api/v1/accounts", tok, accounts.CreateRequest{
		Name: "Mine", Type: "cash", Currency: "COP",
	})
	require.Equal(t, http.StatusCreated, w.Code, "body: %s", w.Body.String())
	var a accounts.Account
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &a))

	// Generate a token for a different user
	otherUserID := uuid.New()
	otherToken, err := auth.GenerateAccessToken(otherUserID, "other@example.com", testSecret)
	require.NoError(t, err)

	// Other user tries to GET — should be 404 because ownership filter
	w = doRequest(r, "GET", "/api/v1/accounts/"+a.ID.String(), otherToken, nil)
	assert.Equal(t, http.StatusNotFound, w.Code)
}