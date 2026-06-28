package auth_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nicolas/finanzas/backend/internal/auth"
	"github.com/nicolas/finanzas/backend/internal/categories"
	"github.com/nicolas/finanzas/backend/internal/config"
	"github.com/nicolas/finanzas/backend/internal/testhelpers"
)

// fakeSeeder records calls and lets us inject errors / counts.
type fakeSeeder struct {
	mu          sync.Mutex
	calls       []uuid.UUID
	returnCount int
	returnErr   error
}

func (f *fakeSeeder) SeedDefaults(userID uuid.UUID) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, userID)
	return f.returnCount, f.returnErr
}

func setupServiceWithSeeder(t *testing.T, seeder auth.CategorySeeder) (auth.Service, *fakeSeeder, func()) {
	t.Helper()
	db := testhelpers.SetupTestDB(t)

	repo := auth.NewUserRepository(db.DB)
	sessions := auth.NewSessionRepository(db.DB)
	cfg := &config.Config{JWTSecret: "test-secret-with-enough-entropy-32-chars-min"}

	base := auth.NewService(repo, sessions, cfg)
	svc := auth.WithCategorySeeder(base, seeder)

	fake := &fakeSeeder{}
	if seeder == nil {
		// Caller didn't wire a seeder — return one anyway for assertion use.
		return svc, fake, db.Cleanup
	}
	// If the caller passed a *fakeSeeder, expose it; otherwise wrap.
	if f, ok := seeder.(*fakeSeeder); ok {
		return svc, f, db.Cleanup
	}
	return svc, fake, db.Cleanup
}

func TestService_Register_SeedsDefaultCategories_WhenSeederWired(t *testing.T) {
	seeder := &fakeSeeder{returnCount: 13}
	svc, fake, cleanup := setupServiceWithSeeder(t, seeder)
	defer cleanup()

	user, err := svc.Register("seedme@example.com", "password123", "Seeder")
	require.NoError(t, err)
	require.NotNil(t, user.ID)

	require.Len(t, fake.calls, 1, "seeder should be called exactly once per registration")
	assert.Equal(t, user.ID, fake.calls[0], "seeder must receive the new user's ID")
}

func TestService_Register_WorksWithoutSeederWired(t *testing.T) {
	// nil seeder is the legacy/dev path — must not break registration.
	svc, fake, cleanup := setupServiceWithSeeder(t, nil)
	defer cleanup()

	user, err := svc.Register("noseed@example.com", "password123", "")
	require.NoError(t, err)
	require.NotNil(t, user.ID)
	assert.Empty(t, fake.calls, "no seeder wired → no calls recorded")
}

func TestService_Register_SeederFailure_DoesNotFailRegistration(t *testing.T) {
	// Best-effort: a flaky seeder must NOT lock the user out of their account.
	seeder := &fakeSeeder{returnErr: errors.New("boom")}
	svc, fake, cleanup := setupServiceWithSeeder(t, seeder)
	defer cleanup()

	user, err := svc.Register("flaky@example.com", "password123", "")
	require.NoError(t, err, "registration must succeed even if seeder fails")
	require.NotNil(t, user.ID)
	assert.Len(t, fake.calls, 1, "seeder was called once before it failed")
}

// Verify the real categories.Service satisfies auth.CategorySeeder — this is
// a compile-time guarantee that the wiring in main.go will work.
func TestCategoriesServiceImplementsAuthSeederInterface(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer db.Cleanup()

	catSvc := categories.NewService(categories.NewCategoryRepository(db.DB))
	var _ auth.CategorySeeder = catSvc
}