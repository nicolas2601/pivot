package budgets

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIsValidPeriod(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"monthly", true},
		{"yearly", true},
		{"weekly", false},
		{"", false},
		{"MONTHLY", false},
	}
	for _, c := range cases {
		if got := IsValidPeriod(c.in); got != c.want {
			t.Errorf("IsValidPeriod(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

type fakeRepo struct {
	mu      sync.Mutex
	budgets map[uuid.UUID]*Budget
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{budgets: map[uuid.UUID]*Budget{}}
}

func (f *fakeRepo) Create(b *Budget) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	copy := *b
	f.budgets[b.ID] = &copy
	return nil
}

func (f *fakeRepo) GetByID(id, userID uuid.UUID) (*Budget, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.budgets[id]
	if !ok || b.UserID != userID {
		return nil, ErrBudgetNotFound
	}
	copy := *b
	return &copy, nil
}

func (f *fakeRepo) ListByUser(userID uuid.UUID) ([]Budget, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Budget{}
	for _, b := range f.budgets {
		if b.UserID == userID {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (f *fakeRepo) Update(b *Budget) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.budgets[b.ID] = b
	return nil
}

func (f *fakeRepo) Delete(id, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	b, ok := f.budgets[id]
	if !ok || b.UserID != userID {
		return ErrBudgetNotFound
	}
	delete(f.budgets, id)
	return nil
}

func (f *fakeRepo) FindByCategoryAndPeriod(userID, categoryID uuid.UUID, period Period) (*Budget, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, b := range f.budgets {
		if b.UserID == userID && b.CategoryID == categoryID && b.Period == period {
			copy := *b
			return &copy, nil
		}
	}
	return nil, ErrBudgetNotFound
}

type fakeCategories struct {
	mu  sync.Mutex
	ids map[uuid.UUID]bool
}

func newFakeCategories() *fakeCategories {
	return &fakeCategories{ids: map[uuid.UUID]bool{}}
}

func (f *fakeCategories) add(id uuid.UUID) { f.ids[id] = true }

func (f *fakeCategories) GetByID(id, _ uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ids[id], nil
}

func validCreateReq() CreateRequest {
	return CreateRequest{
		CategoryID: uuid.New().String(),
		Amount:     500000,
		Period:     "monthly",
		StartDate:  "2026-01-01",
	}
}

func TestService_Create_HappyPath(t *testing.T) {
	repo := newFakeRepo()
	cat := newFakeCategories()
	catID := uuid.New()
	cat.add(catID)
	s := NewService(repo, cat)

	req := validCreateReq()
	req.CategoryID = catID.String()
	b, err := s.Create(uuid.New(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if b.Amount != 500000 {
		t.Errorf("Amount = %d, want 500000", b.Amount)
	}
	if b.Period != PeriodMonthly {
		t.Errorf("Period = %v, want PeriodMonthly", b.Period)
	}
}

func TestService_Create_RejectsInvalidPeriod(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeCategories())
	req := validCreateReq()
	req.Period = "weekly"
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidPeriod) {
		t.Errorf("got %v, want ErrInvalidPeriod", err)
	}
}

func TestService_Create_RejectsNonPositiveAmount(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeCategories())
	for _, amt := range []int64{0, -1} {
		req := validCreateReq()
		req.Amount = amt
		if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("amount=%d: got %v, want ErrInvalidAmount", amt, err)
		}
	}
}

func TestService_Create_RejectsEmptyStartDate(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeCategories())
	req := validCreateReq()
	req.StartDate = ""
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidDate) {
		t.Errorf("got %v, want ErrInvalidDate", err)
	}
}

func TestService_Create_RejectsEndBeforeStart(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeCategories())
	end := "2025-01-01"
	req := validCreateReq()
	req.StartDate = "2026-01-01"
	req.EndDate = &end
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrEndBeforeStart) {
		t.Errorf("got %v, want ErrEndBeforeStart", err)
	}
}

func TestService_Create_RejectsUnknownCategory(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeCategories()) // empty cat
	if _, err := s.Create(uuid.New(), validCreateReq()); !errors.Is(err, ErrCategoryMissing) {
		t.Errorf("got %v, want ErrCategoryMissing", err)
	}
}

func TestService_Update_CanClearEndDate(t *testing.T) {
	repo := newFakeRepo()
	cat := newFakeCategories()
	catID := uuid.New()
	cat.add(catID)
	s := NewService(repo, cat)

	req := validCreateReq()
	req.CategoryID = catID.String()
	end := "2026-12-31"
	req.EndDate = &end
	b, err := s.Create(uuid.New(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if b.EndDate == nil {
		t.Fatalf("precondition: EndDate should be set")
	}

	updated, err := s.Update(b.ID, b.UserID, UpdateRequest{ClearEnd: true})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.EndDate != nil {
		t.Errorf("EndDate = %v, want nil after clear", updated.EndDate)
	}
}

func TestService_Update_AcceptsNewPeriod(t *testing.T) {
	repo := newFakeRepo()
	cat := newFakeCategories()
	catID := uuid.New()
	cat.add(catID)
	s := NewService(repo, cat)

	req := validCreateReq()
	req.CategoryID = catID.String()
	b, err := s.Create(uuid.New(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := s.Update(b.ID, b.UserID, UpdateRequest{Period: stringPtr("yearly")})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Period != PeriodYearly {
		t.Errorf("Period = %v, want PeriodYearly", updated.Period)
	}
}

func TestService_Delete_OwnershipEnforced(t *testing.T) {
	repo := newFakeRepo()
	cat := newFakeCategories()
	catID := uuid.New()
	cat.add(catID)
	s := NewService(repo, cat)

	owner := uuid.New()
	other := uuid.New()
	req := validCreateReq()
	req.CategoryID = catID.String()
	b, err := s.Create(owner, req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := s.Delete(b.ID, other); !errors.Is(err, ErrBudgetNotFound) {
		t.Errorf("other delete: got %v, want ErrBudgetNotFound", err)
	}
	if err := s.Delete(b.ID, owner); err != nil {
		t.Errorf("owner delete: %v", err)
	}
}

func stringPtr(s string) *string { return &s }

// sanity: ensure time package is referenced even when nothing else uses it
var _ = time.Now