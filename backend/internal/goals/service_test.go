package goals

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeRepo implements Repository for service-level tests.
type fakeRepo struct {
	mu    sync.Mutex
	goals map[uuid.UUID]*Goal
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{goals: map[uuid.UUID]*Goal{}}
}

func (f *fakeRepo) Create(g *Goal) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	copy := *g
	f.goals[g.ID] = &copy
	return nil
}

func (f *fakeRepo) GetByID(id, userID uuid.UUID) (*Goal, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	g, ok := f.goals[id]
	if !ok || g.UserID != userID {
		return nil, ErrGoalNotFound
	}
	copy := *g
	return &copy, nil
}

func (f *fakeRepo) ListByUser(userID uuid.UUID) ([]Goal, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Goal{}
	for _, g := range f.goals {
		if g.UserID == userID {
			out = append(out, *g)
		}
	}
	return out, nil
}

func (f *fakeRepo) Update(g *Goal) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.goals[g.ID]
	if !ok {
		return ErrGoalNotFound
	}
	*existing = *g
	return nil
}

func (f *fakeRepo) Delete(id, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	g, ok := f.goals[id]
	if !ok || g.UserID != userID {
		return ErrGoalNotFound
	}
	delete(f.goals, id)
	return nil
}

func (f *fakeRepo) AtomicAdjustCurrent(id, userID uuid.UUID, delta int64) (*Goal, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	g, ok := f.goals[id]
	if !ok || g.UserID != userID {
		return nil, ErrGoalNotFound
	}
	g.CurrentAmount += delta
	now := time.Now()
	if g.CurrentAmount >= g.TargetAmount && !g.IsCompleted {
		g.IsCompleted = true
		g.CompletedAt = &now
	} else if g.CurrentAmount < g.TargetAmount && g.IsCompleted {
		g.IsCompleted = false
		g.CompletedAt = nil
	}
	copy := *g
	return &copy, nil
}

// fakeAccountLookup implements AccountLookup with a configurable map.
type fakeAccountLookup struct {
	mu       sync.Mutex
	accounts map[uuid.UUID]bool
	err      error
}

func newFakeAccountLookup() *fakeAccountLookup {
	return &fakeAccountLookup{accounts: map[uuid.UUID]bool{}}
}

func (f *fakeAccountLookup) add(id uuid.UUID) { f.accounts[id] = true }

func (f *fakeAccountLookup) Exists(id, _ uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return false, f.err
	}
	return f.accounts[id], nil
}

func validCreateReq() CreateRequest {
	deadline := "2030-12-31"
	return CreateRequest{
		Name:         "Vacaciones Europa",
		TargetAmount: 5000000,
		Currency:     "COP",
		Deadline:     &deadline,
	}
}

func TestService_Create_HappyPath(t *testing.T) {
	repo := newFakeRepo()
	accounts := newFakeAccountLookup()
	s := NewService(repo, accounts)

	goal, err := s.Create(uuid.New(), validCreateReq())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if goal.Name != "Vacaciones Europa" {
		t.Errorf("Name = %q, want Vacaciones Europa", goal.Name)
	}
	if goal.TargetAmount != 5000000 {
		t.Errorf("TargetAmount = %d, want 5000000", goal.TargetAmount)
	}
	if goal.CurrentAmount != 0 {
		t.Errorf("CurrentAmount = %d, want 0 (default)", goal.CurrentAmount)
	}
	if goal.IsCompleted {
		t.Errorf("IsCompleted = true, want false")
	}
	if goal.Currency != "COP" {
		t.Errorf("Currency = %q, want COP", goal.Currency)
	}
}

func TestService_Create_DefaultsCurrencyToCOP(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeAccountLookup())

	req := validCreateReq()
	req.Currency = ""
	goal, err := s.Create(uuid.New(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if goal.Currency != "COP" {
		t.Errorf("Currency = %q, want COP (default)", goal.Currency)
	}
}

func TestService_Create_RejectsEmptyName(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	req := validCreateReq()
	req.Name = ""
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidName) {
		t.Errorf("got %v, want ErrInvalidName", err)
	}
}

func TestService_Create_RejectsNonPositiveAmount(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	for _, amt := range []int64{0, -1, -1000} {
		req := validCreateReq()
		req.TargetAmount = amt
		if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("amount=%d: got %v, want ErrInvalidAmount", amt, err)
		}
	}
}

func TestService_Create_RejectsInvalidCurrency(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	req := validCreateReq()
	req.Currency = "PESO"
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidCurrency) {
		t.Errorf("got %v, want ErrInvalidCurrency", err)
	}
}

func TestService_Create_RejectsPastDeadline(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	past := "2020-01-01"
	req := validCreateReq()
	req.Deadline = &past
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidDeadline) {
		t.Errorf("got %v, want ErrInvalidDeadline", err)
	}
}

func TestService_Create_RejectsBadDateFormat(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	bad := "not-a-date"
	req := validCreateReq()
	req.Deadline = &bad
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidDeadline) {
		t.Errorf("got %v, want ErrInvalidDeadline", err)
	}
}

func TestService_Create_AcceptsRFC3339Deadline(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	iso := "2030-12-31T00:00:00Z"
	req := validCreateReq()
	req.Deadline = &iso
	goal, err := s.Create(uuid.New(), req)
	if err != nil {
		t.Fatalf("Create with RFC3339 deadline: %v", err)
	}
	if goal.Deadline == nil {
		t.Fatalf("Deadline nil")
	}
	if goal.Deadline.Year() != 2030 {
		t.Errorf("year = %d, want 2030", goal.Deadline.Year())
	}
}

func TestService_Create_RejectsUnknownAccount(t *testing.T) {
	accounts := newFakeAccountLookup() // empty
	s := NewService(newFakeRepo(), accounts)

	accID := uuid.New().String()
	req := validCreateReq()
	req.AccountID = &accID
	if _, err := s.Create(uuid.New(), req); err == nil || !strings.Contains(err.Error(), "account not found") {
		t.Errorf("got %v, want account-not-found", err)
	}
}

func TestService_Deposit_AddsToCurrent(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeAccountLookup())
	goal, err := s.Create(uuid.New(), validCreateReq())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := s.Deposit(goal.ID, goal.UserID, MoveRequest{Amount: 100000})
	if err != nil {
		t.Fatalf("Deposit: %v", err)
	}
	if updated.CurrentAmount != 100000 {
		t.Errorf("CurrentAmount = %d, want 100000", updated.CurrentAmount)
	}
	if updated.IsCompleted {
		t.Errorf("IsCompleted = true, want false (still under target)")
	}
}

func TestService_Deposit_ReachesTargetMarksCompleted(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeAccountLookup())
	goal, err := s.Create(uuid.New(), validCreateReq())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Deposit full amount — should mark completed.
	updated, err := s.Deposit(goal.ID, goal.UserID, MoveRequest{Amount: goal.TargetAmount})
	if err != nil {
		t.Fatalf("Deposit: %v", err)
	}
	if !updated.IsCompleted {
		t.Errorf("IsCompleted = false, want true")
	}
	if updated.CompletedAt == nil {
		t.Errorf("CompletedAt nil, want set")
	}
}

func TestService_Deposit_RejectsNonPositive(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	goal, _ := s.Create(uuid.New(), validCreateReq())

	for _, amt := range []int64{0, -1} {
		if _, err := s.Deposit(goal.ID, goal.UserID, MoveRequest{Amount: amt}); !errors.Is(err, ErrNonPositiveMove) {
			t.Errorf("amount=%d: got %v, want ErrNonPositiveMove", amt, err)
		}
	}
}

func TestService_Withdraw_SubtractsFromCurrent(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeAccountLookup())
	goal, _ := s.Create(uuid.New(), validCreateReq())
	_, _ = s.Deposit(goal.ID, goal.UserID, MoveRequest{Amount: 200000})

	updated, err := s.Withdraw(goal.ID, goal.UserID, MoveRequest{Amount: 50000})
	if err != nil {
		t.Fatalf("Withdraw: %v", err)
	}
	if updated.CurrentAmount != 150000 {
		t.Errorf("CurrentAmount = %d, want 150000", updated.CurrentAmount)
	}
}

func TestService_Withdraw_RejectsOverdraw(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeAccountLookup())
	goal, _ := s.Create(uuid.New(), validCreateReq())
	_, _ = s.Deposit(goal.ID, goal.UserID, MoveRequest{Amount: 50000})

	if _, err := s.Withdraw(goal.ID, goal.UserID, MoveRequest{Amount: 50001}); !errors.Is(err, ErrOverWithdraw) {
		t.Errorf("got %v, want ErrOverWithdraw", err)
	}
}

func TestService_Withdraw_RejectsNonPositive(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeAccountLookup())
	goal, _ := s.Create(uuid.New(), validCreateReq())

	for _, amt := range []int64{0, -100} {
		if _, err := s.Withdraw(goal.ID, goal.UserID, MoveRequest{Amount: amt}); !errors.Is(err, ErrNonPositiveMove) {
			t.Errorf("amount=%d: got %v, want ErrNonPositiveMove", amt, err)
		}
	}
}

func TestService_Update_CanClearDeadline(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeAccountLookup())
	goal, _ := s.Create(uuid.New(), validCreateReq())
	if goal.Deadline == nil {
		t.Fatalf("precondition: Deadline should be set")
	}

	req := UpdateRequest{ClearDeadline: true}
	updated, err := s.Update(goal.ID, goal.UserID, req)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Deadline != nil {
		t.Errorf("Deadline = %v, want nil after clear", updated.Deadline)
	}
}

func TestService_Update_CanClearAccount(t *testing.T) {
	repo := newFakeRepo()
	accounts := newFakeAccountLookup()
	accID := uuid.New()
	accounts.add(accID)
	s := NewService(repo, accounts)

	goal, _ := s.Create(uuid.New(), validCreateReq())
	goal.AccountID = &accID
	if err := repo.Update(goal); err != nil {
		t.Fatalf("seed: %v", err)
	}

	req := UpdateRequest{ClearAccount: true}
	updated, err := s.Update(goal.ID, goal.UserID, req)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.AccountID != nil {
		t.Errorf("AccountID = %v, want nil after clear", updated.AccountID)
	}
}

func TestService_Delete_OwnershipEnforced(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeAccountLookup())

	owner := uuid.New()
	other := uuid.New()
	goal, _ := s.Create(owner, validCreateReq())

	if err := s.Delete(goal.ID, other); !errors.Is(err, ErrGoalNotFound) {
		t.Errorf("other delete: got %v, want ErrGoalNotFound", err)
	}
	if err := s.Delete(goal.ID, owner); err != nil {
		t.Errorf("owner delete: %v", err)
	}
	if _, err := repo.GetByID(goal.ID, owner); !errors.Is(err, ErrGoalNotFound) {
		t.Errorf("post-delete: got %v, want ErrGoalNotFound", err)
	}
}