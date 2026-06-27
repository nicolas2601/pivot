package accounts_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicolas/finanzas/backend/internal/accounts"
	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/testhelpers"
)

func setupAccountRepo(t *testing.T) (accounts.AccountRepository, uuid.UUID, func()) {
	t.Helper()
	db := testhelpers.SetupTestDB(t)
	userRepo := auth.NewUserRepository(db.DB)
	email := uuid.NewString() + "@example.com"
	err := userRepo.Create(&auth.User{
		Email:        email,
		PasswordHash: "x",
	})
	require.NoError(t, err)
	user, err := userRepo.FindByEmail(email)
	require.NoError(t, err)
	return accounts.NewAccountRepository(db.DB), user.ID, db.Cleanup
}

func TestAccountRepository_CreateAndList(t *testing.T) {
	repo, userID, cleanup := setupAccountRepo(t)
	defer cleanup()

	acc := &accounts.Account{
		UserID:   userID,
		Name:     "Bancolombia",
		Type:     accounts.TypeDebit,
		Currency: "COP",
	}
	require.NoError(t, repo.Create(acc))

	list, err := repo.ListByUser(userID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "Bancolombia", list[0].Name)
	assert.NotEqual(t, uuid.Nil, list[0].ID)
}

func TestAccountRepository_GetByID_OwnershipEnforced(t *testing.T) {
	repo, userID, cleanup := setupAccountRepo(t)
	defer cleanup()

	acc := &accounts.Account{UserID: userID, Name: "A", Type: accounts.TypeCash, Currency: "COP"}
	require.NoError(t, repo.Create(acc))

	found, err := repo.GetByID(acc.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "A", found.Name)

	_, err = repo.GetByID(acc.ID, uuid.New())
	assert.ErrorIs(t, err, accounts.ErrAccountNotFound)
}

func TestAccountRepository_SoftDelete(t *testing.T) {
	repo, userID, cleanup := setupAccountRepo(t)
	defer cleanup()

	acc := &accounts.Account{UserID: userID, Name: "X", Type: accounts.TypeCash, Currency: "COP"}
	require.NoError(t, repo.Create(acc))

	require.NoError(t, repo.Delete(acc.ID, userID))

	list, err := repo.ListByUser(userID)
	require.NoError(t, err)
	assert.Empty(t, list)

	_, err = repo.GetByID(acc.ID, userID)
	assert.ErrorIs(t, err, accounts.ErrAccountNotFound)
}

func TestAccountRepository_Update(t *testing.T) {
	repo, userID, cleanup := setupAccountRepo(t)
	defer cleanup()

	acc := &accounts.Account{UserID: userID, Name: "Old", Type: accounts.TypeCash, Currency: "COP"}
	require.NoError(t, repo.Create(acc))

	acc.Name = "New"
	require.NoError(t, repo.Update(acc))

	found, err := repo.GetByID(acc.ID, userID)
	require.NoError(t, err)
	assert.Equal(t, "New", found.Name)
}