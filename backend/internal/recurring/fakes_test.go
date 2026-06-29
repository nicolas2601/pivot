package recurring

import (
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// fakeRepo implements Repository for service-level tests.
type fakeRepo struct {
	mu    sync.Mutex
	rules map[uuid.UUID]*Rule
	runs  map[uuid.UUID]*Run
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		rules: map[uuid.UUID]*Rule{},
		runs:  map[uuid.UUID]*Run{},
	}
}

func (f *fakeRepo) CreateRule(r *Rule) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	copy := *r
	f.rules[r.ID] = &copy
	return nil
}

func (f *fakeRepo) GetRuleByID(id, userID uuid.UUID) (*Rule, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rules[id]
	if !ok || r.UserID != userID {
		return nil, ErrRuleNotFound
	}
	copy := *r
	return &copy, nil
}

func (f *fakeRepo) ListRulesByUser(userID uuid.UUID) ([]Rule, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Rule{}
	for _, r := range f.rules {
		if r.UserID == userID {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeRepo) UpdateRule(r *Rule) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.rules[r.ID]
	if !ok {
		return ErrRuleNotFound
	}
	*existing = *r
	return nil
}

func (f *fakeRepo) DeleteRule(id, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rules[id]
	if !ok || r.UserID != userID {
		return ErrRuleNotFound
	}
	delete(f.rules, id)
	return nil
}

func (f *fakeRepo) ListDueRules(before time.Time) ([]Rule, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Rule{}
	for _, r := range f.rules {
		if r.IsActive && !r.NextRunDate.After(before) {
			out = append(out, *r)
		}
	}
	return out, nil
}

func (f *fakeRepo) CreateRun(run *Run) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if run.ID == uuid.Nil {
		run.ID = uuid.New()
	}
	// UNIQUE(rule_id, scheduled_date).
	for _, existing := range f.runs {
		if existing.RecurringRuleID == run.RecurringRuleID && existing.ScheduledDate.Equal(run.ScheduledDate) {
			return errors.New("unique constraint violation")
		}
	}
	copy := *run
	f.runs[run.ID] = &copy
	return nil
}

func (f *fakeRepo) GetRunByRuleAndDate(ruleID uuid.UUID, scheduled time.Time) (*Run, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, r := range f.runs {
		if r.RecurringRuleID == ruleID && r.ScheduledDate.Equal(scheduled) {
			copy := *r
			return &copy, nil
		}
	}
	return nil, ErrRunNotFound
}

func (f *fakeRepo) ListRunsByRule(ruleID uuid.UUID, limit int) ([]Run, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Run{}
	for _, r := range f.runs {
		if r.RecurringRuleID == ruleID {
			out = append(out, *r)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (f *fakeRepo) UpdateRun(run *Run) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.runs[run.ID]
	if !ok {
		return ErrRunNotFound
	}
	*existing = *run
	return nil
}

// fakeAccountLookup implements AccountLookup with a configurable map of IDs.
type fakeAccountLookup struct {
	mu       sync.Mutex
	accounts map[uuid.UUID]bool
	err      error
}

func newFakeAccountLookup() *fakeAccountLookup {
	return &fakeAccountLookup{accounts: map[uuid.UUID]bool{}}
}

func (f *fakeAccountLookup) addAccount(id uuid.UUID) { f.accounts[id] = true }

func (f *fakeAccountLookup) Exists(id, _ uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return false, f.err
	}
	return f.accounts[id], nil
}

// fakeCategoryLookup implements CategoryLookup with a configurable map of IDs.
type fakeCategoryLookup struct {
	mu         sync.Mutex
	categories map[uuid.UUID]bool
	err        error
}

func newFakeCategoryLookup() *fakeCategoryLookup {
	return &fakeCategoryLookup{categories: map[uuid.UUID]bool{}}
}

func (f *fakeCategoryLookup) addCategory(id uuid.UUID) { f.categories[id] = true }

func (f *fakeCategoryLookup) Exists(id, _ uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return false, f.err
	}
	return f.categories[id], nil
}

// fakeTxCreator records calls and returns a configurable tx ID or error.
type fakeTxCreator struct {
	mu      sync.Mutex
	calls   []txCreatorCall
	txID    uuid.UUID
	err     error
}

type txCreatorCall struct {
	UserID    uuid.UUID
	AccountID uuid.UUID
	Category  uuid.UUID
	Type      string
	Amount    int64
	Currency  string
	Date      time.Time
	RunID     uuid.UUID
}

func newFakeTxCreator() *fakeTxCreator {
	return &fakeTxCreator{txID: uuid.New()}
}

func (f *fakeTxCreator) CreateFromRecurring(userID, accountID, categoryID uuid.UUID,
	txType string, amount int64, currency string, date time.Time,
	description, notes *string, recurringRunID uuid.UUID,
) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, txCreatorCall{
		UserID: userID, AccountID: accountID, Category: categoryID,
		Type: txType, Amount: amount, Currency: currency, Date: date,
		RunID: recurringRunID,
	})
	if f.err != nil {
		return uuid.Nil, f.err
	}
	return f.txID, nil
}

// fakeUserResolver is trivial: it just parses the UUID string.
type fakeUserResolver struct{}

func (fakeUserResolver) UserIDFromString(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// newTestService wires up Service with fake dependencies and an optional
// clock for deterministic "today".
func newTestService(repo Repository, accounts AccountLookup, categories AccountLookup,
	txCreator TxCreator, users UserResolver, now time.Time,
) *Service {
	s := NewService(repo, accounts, categories, txCreator, users)
	s.now = func() time.Time { return now }
	return s
}