package transactions

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIsValidType(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"expense", true},
		{"income", true},
		{"transfer", true},
		{"", false},
		{"EXPENSE", false}, // case-sensitive on purpose
		{"savings", false},
	}
	for _, c := range cases {
		got := IsValidType(c.in)
		if got != c.want {
			t.Errorf("IsValidType(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// fakeAccount implements Account for service tests.
type fakeAccount struct {
	mu        sync.Mutex
	accounts  map[uuid.UUID]AccountInfo
	errOnGet  error
}

func newFakeAccount() *fakeAccount {
	return &fakeAccount{accounts: map[uuid.UUID]AccountInfo{}}
}

func (f *fakeAccount) add(id, userID uuid.UUID, currency string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.accounts[id] = AccountInfo{ID: id, UserID: userID, Currency: currency}
}

func (f *fakeAccount) GetByID(id, userID uuid.UUID) (AccountInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.errOnGet != nil {
		return AccountInfo{}, f.errOnGet
	}
	a, ok := f.accounts[id]
	if !ok || a.UserID != userID {
		return AccountInfo{}, errors.New("account not found")
	}
	return a, nil
}

// fakeCategory implements CategoryLookup with a configurable map.
type fakeCategory struct {
	mu        sync.Mutex
	categories map[uuid.UUID]bool
	err       error
}

func newFakeCategory() *fakeCategory {
	return &fakeCategory{categories: map[uuid.UUID]bool{}}
}

func (f *fakeCategory) add(id uuid.UUID) { f.categories[id] = true }

func (f *fakeCategory) GetByID(id, _ uuid.UUID) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return false, f.err
	}
	return f.categories[id], nil
}

// fakeTxRepo implements Repository for service tests.
type fakeTxRepo struct {
	mu          sync.Mutex
	transactions map[uuid.UUID]*Transaction
	pairs       map[uuid.UUID]bool
	createErr   error
	transferErr error
}

func newFakeTxRepo() *fakeTxRepo {
	return &fakeTxRepo{
		transactions: map[uuid.UUID]*Transaction{},
		pairs:       map[uuid.UUID]bool{},
	}
}

func (f *fakeTxRepo) Create(tx *Transaction) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.createErr != nil {
		return f.createErr
	}
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	copy := *tx
	f.transactions[tx.ID] = &copy
	return nil
}

func (f *fakeTxRepo) GetByID(id, userID uuid.UUID) (*Transaction, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.transactions[id]
	if !ok || t.UserID != userID || t.DeletedAt != nil {
		return nil, ErrTransactionNotFound
	}
	copy := *t
	return &copy, nil
}

func (f *fakeTxRepo) ListByUser(_ uuid.UUID, _ ListFilter) ([]Transaction, error) {
	return nil, nil
}

func (f *fakeTxRepo) Update(tx *Transaction) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	existing, ok := f.transactions[tx.ID]
	if !ok {
		return ErrTransactionNotFound
	}
	*existing = *tx
	return nil
}

func (f *fakeTxRepo) Delete(id, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	t, ok := f.transactions[id]
	if !ok || t.UserID != userID {
		return ErrTransactionNotFound
	}
	now := time.Now()
	t.DeletedAt = &now
	return nil
}

func (f *fakeTxRepo) CreateTransfer(userID uuid.UUID, source, dest *Transaction) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.transferErr != nil {
		return f.transferErr
	}
	if source.AccountID == dest.AccountID {
		return ErrAccountMismatch
	}
	if source.Currency != dest.Currency {
		return ErrCurrencyMismatch
	}
	pairID := uuid.New()
	source.TransferPairID = &pairID
	dest.TransferPairID = &pairID
	if source.ID == uuid.Nil {
		source.ID = uuid.New()
	}
	if dest.ID == uuid.Nil {
		dest.ID = uuid.New()
	}
	f.transactions[source.ID] = &Transaction{}
	*f.transactions[source.ID] = *source
	f.transactions[dest.ID] = &Transaction{}
	*f.transactions[dest.ID] = *dest
	f.pairs[pairID] = true
	return nil
}

func (f *fakeTxRepo) DeletePair(pairID, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for id, t := range f.transactions {
		if t.TransferPairID != nil && *t.TransferPairID == pairID && t.UserID == userID {
			now := time.Now()
			t.DeletedAt = &now
			f.transactions[id] = t
		}
	}
	delete(f.pairs, pairID)
	return nil
}

func (f *fakeTxRepo) SumByCategory(_ uuid.UUID, _, _ time.Time) ([]CategorySum, error) {
	return nil, nil
}

func (f *fakeTxRepo) SumByAccount(_ uuid.UUID, _, _ time.Time) ([]AccountSum, error) {
	return nil, nil
}

func (f *fakeTxRepo) MonthlyTrend(_ uuid.UUID, _, _ time.Time) ([]MonthlyTotal, error) {
	return nil, nil
}

// ──────────────────────────── Service tests ────────────────────────────

func validCreateReq(accountID uuid.UUID) CreateRequest {
	return CreateRequest{
		AccountID: accountID.String(),
		Type:      "expense",
		Amount:    5000,
		Currency:  "COP",
		Date:      "2026-01-15",
	}
}

func TestService_Create_HappyPath(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	cat := newFakeCategory()
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	acc.add(accID, userID, "COP")
	cat.add(catID)

	s := NewService(repo, acc, cat)
	catStr := catID.String()
	req := validCreateReq(accID)
	req.CategoryID = &catStr
	tx, err := s.Create(userID, req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tx.Amount != 5000 {
		t.Errorf("Amount = %d, want 5000", tx.Amount)
	}
	if tx.Type != TypeExpense {
		t.Errorf("Type = %v, want expense", tx.Type)
	}
	if tx.CategoryID == nil || *tx.CategoryID != catID {
		t.Errorf("CategoryID = %v, want %v", tx.CategoryID, catID)
	}
}

func TestService_Create_RejectsInvalidType(t *testing.T) {
	s := NewService(newFakeTxRepo(), newFakeAccount(), newFakeCategory())
	req := validCreateReq(uuid.New())
	req.Type = "savings"
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidType) {
		t.Errorf("got %v, want ErrInvalidType", err)
	}
}

func TestService_Create_RejectsTransferViaCreate(t *testing.T) {
	s := NewService(newFakeTxRepo(), newFakeAccount(), newFakeCategory())
	req := validCreateReq(uuid.New())
	req.Type = "transfer" // must use Transfer(), not Create()
	_, err := s.Create(uuid.New(), req)
	if err == nil || !contains(err.Error(), "use Transfer()") {
		t.Errorf("got %v, want use-Transfer() error", err)
	}
}

func TestService_Create_RejectsNonPositiveAmount(t *testing.T) {
	s := NewService(newFakeTxRepo(), newFakeAccount(), newFakeCategory())
	for _, amt := range []int64{0, -1} {
		req := validCreateReq(uuid.New())
		req.Amount = amt
		if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("amount=%d: got %v, want ErrInvalidAmount", amt, err)
		}
	}
}

func TestService_Create_RejectsUnknownAccount(t *testing.T) {
	acc := newFakeAccount() // empty
	s := NewService(newFakeTxRepo(), acc, newFakeCategory())
	if _, err := s.Create(uuid.New(), validCreateReq(uuid.New())); !errors.Is(err, ErrAccountNotFound) {
		t.Errorf("got %v, want ErrAccountNotFound", err)
	}
}

func TestService_Create_RejectsUnknownCategory(t *testing.T) {
	acc := newFakeAccount()
	cat := newFakeCategory() // empty
	userID := uuid.New()
	accID := uuid.New()
	acc.add(accID, userID, "COP")
	s := NewService(newFakeTxRepo(), acc, cat)

	catStr := uuid.New().String()
	req := validCreateReq(accID)
	req.CategoryID = &catStr
	if _, err := s.Create(userID, req); !errors.Is(err, ErrCategoryNotFound) {
		t.Errorf("got %v, want ErrCategoryNotFound", err)
	}
}

func TestService_Create_DefaultsCurrencyFromAccount(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	userID := uuid.New()
	accID := uuid.New()
	acc.add(accID, userID, "USD")
	s := NewService(repo, acc, newFakeCategory())

	req := validCreateReq(accID)
	req.Currency = "" // should default to account currency
	tx, err := s.Create(userID, req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tx.Currency != "USD" {
		t.Errorf("Currency = %q, want USD (from account)", tx.Currency)
	}
}

func TestService_Transfer_HappyPath(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	userID := uuid.New()
	fromID, toID := uuid.New(), uuid.New()
	acc.add(fromID, userID, "COP")
	acc.add(toID, userID, "COP")
	s := NewService(repo, acc, newFakeCategory())

	res, err := s.Transfer(userID, TransferRequest{
		FromAccountID: fromID.String(),
		ToAccountID:   toID.String(),
		Amount:        10000,
		Currency:      "COP",
		Date:          "2026-01-15",
	})
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}
	if res.Source == nil || res.Dest == nil {
		t.Fatalf("nil tx in result: %+v", res)
	}
	if res.Source.Type != TypeTransfer || res.Dest.Type != TypeTransfer {
		t.Errorf("both legs must be type=transfer")
	}
	if res.Source.TransferPairID == nil || res.Dest.TransferPairID == nil {
		t.Fatalf("TransferPairID not set")
	}
	if *res.Source.TransferPairID != *res.Dest.TransferPairID {
		t.Errorf("TransferPairID mismatch: source=%v dest=%v", res.Source.TransferPairID, res.Dest.TransferPairID)
	}
	if res.Source.AccountID != fromID {
		t.Errorf("source account = %v, want %v", res.Source.AccountID, fromID)
	}
	if res.Dest.AccountID != toID {
		t.Errorf("dest account = %v, want %v", res.Dest.AccountID, toID)
	}
}

func TestService_Transfer_RejectsSameAccount(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	userID := uuid.New()
	accID := uuid.New()
	acc.add(accID, userID, "COP")
	s := NewService(repo, acc, newFakeCategory())

	_, err := s.Transfer(userID, TransferRequest{
		FromAccountID: accID.String(),
		ToAccountID:   accID.String(),
		Amount:        100,
		Currency:      "COP",
		Date:          "2026-01-15",
	})
	if !errors.Is(err, ErrAccountMismatch) {
		t.Errorf("got %v, want ErrAccountMismatch", err)
	}
}

func TestService_Transfer_RejectsCurrencyMismatch(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	userID := uuid.New()
	fromID, toID := uuid.New(), uuid.New()
	acc.add(fromID, userID, "COP")
	acc.add(toID, userID, "USD")
	s := NewService(repo, acc, newFakeCategory())

	_, err := s.Transfer(userID, TransferRequest{
		FromAccountID: fromID.String(),
		ToAccountID:   toID.String(),
		Amount:        100,
		Currency:      "COP",
		Date:          "2026-01-15",
	})
	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Errorf("got %v, want ErrCurrencyMismatch", err)
	}
}

func TestService_Transfer_RejectsNonPositiveAmount(t *testing.T) {
	s := NewService(newFakeTxRepo(), newFakeAccount(), newFakeCategory())
	for _, amt := range []int64{0, -1, -100} {
		_, err := s.Transfer(uuid.New(), TransferRequest{
			FromAccountID: uuid.New().String(),
			ToAccountID:   uuid.New().String(),
			Amount:        amt,
			Currency:      "COP",
			Date:          "2026-01-15",
		})
		if !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("amount=%d: got %v, want ErrInvalidAmount", amt, err)
		}
	}
}

func TestService_Update_RejectsTransferMutation(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	userID := uuid.New()
	fromID, toID := uuid.New(), uuid.New()
	acc.add(fromID, userID, "COP")
	acc.add(toID, userID, "COP")
	s := NewService(repo, acc, newFakeCategory())

	res, err := s.Transfer(userID, TransferRequest{
		FromAccountID: fromID.String(),
		ToAccountID:   toID.String(),
		Amount:        100,
		Currency:      "COP",
		Date:          "2026-01-15",
	})
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}

	newAmt := int64(200)
	_, err = s.Update(res.Source.ID, userID, UpdateRequest{Amount: &newAmt})
	if err == nil || !contains(err.Error(), "immutable") {
		t.Errorf("got %v, want immutable error", err)
	}
}

func TestService_Delete_TransferCascadesPair(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	userID := uuid.New()
	fromID, toID := uuid.New(), uuid.New()
	acc.add(fromID, userID, "COP")
	acc.add(toID, userID, "COP")
	s := NewService(repo, acc, newFakeCategory())

	res, err := s.Transfer(userID, TransferRequest{
		FromAccountID: fromID.String(),
		ToAccountID:   toID.String(),
		Amount:        100,
		Currency:      "COP",
		Date:          "2026-01-15",
	})
	if err != nil {
		t.Fatalf("Transfer: %v", err)
	}

	// Delete source leg → both should be soft-deleted. GetByID treats
	// soft-deleted as not-found, so check internal repo state directly.
	if err := s.Delete(res.Source.ID, userID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.transactions[res.Source.ID].DeletedAt == nil {
		t.Errorf("source not soft-deleted")
	}
	if repo.transactions[res.Dest.ID].DeletedAt == nil {
		t.Errorf("dest not soft-deleted")
	}
}

func TestService_CreateFromRecurring_RejectsTransfer(t *testing.T) {
	s := NewService(newFakeTxRepo(), newFakeAccount(), newFakeCategory())
	_, err := s.CreateFromRecurring(
		uuid.New(), uuid.New(), uuid.New(),
		"transfer", 100, "COP", time.Now(),
		nil, nil, uuid.New(),
	)
	if !errors.Is(err, ErrInvalidType) {
		t.Errorf("got %v, want ErrInvalidType", err)
	}
}

func TestService_CreateFromRecurring_HappyPath(t *testing.T) {
	repo := newFakeTxRepo()
	acc := newFakeAccount()
	cat := newFakeCategory()
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	acc.add(accID, userID, "COP")
	cat.add(catID)
	s := NewService(repo, acc, cat)

	runID := uuid.New()
	txID, err := s.CreateFromRecurring(
		userID, accID, catID,
		"expense", 9900, "COP", time.Now(),
		nil, nil, runID,
	)
	if err != nil {
		t.Fatalf("CreateFromRecurring: %v", err)
	}
	if txID == uuid.Nil {
		t.Fatalf("txID is nil")
	}
	stored, _ := repo.GetByID(txID, userID)
	if stored == nil {
		t.Fatalf("stored tx not found")
	}
	if stored.Amount != 9900 {
		t.Errorf("Amount = %d, want 9900", stored.Amount)
	}
}

// tiny helper to avoid pulling strings.Contains everywhere.
func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}