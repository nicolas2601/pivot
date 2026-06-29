package recurring

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validCreateReq(accountID, categoryID uuid.UUID) CreateRequest {
	return CreateRequest{
		AccountID:     accountID.String(),
		CategoryID:    categoryID.String(),
		Type:          "expense",
		Amount:        10000,
		Currency:      "COP",
		Frequency:     "monthly",
		IntervalCount: 1,
		StartDate:     "2026-01-01",
	}
}

func TestService_Create_HappyPath(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)

	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, newFakeTxCreator(), fakeUserResolver{}, today)

	rule, err := s.Create(userID, validCreateReq(accID, catID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rule.UserID != userID {
		t.Errorf("UserID = %v, want %v", rule.UserID, userID)
	}
	// Next run is the first strictly-after today, i.e. Feb 1.
	want := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	if !rule.NextRunDate.Equal(want) {
		t.Errorf("NextRunDate = %v, want %v", rule.NextRunDate, want)
	}
	if !rule.IsActive {
		t.Errorf("IsActive = false, want true")
	}
}

func TestService_Create_RejectsInvalidAmount(t *testing.T) {
	repo := newFakeRepo()
	accID := uuid.New()
	catID := uuid.New()
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, newFakeAccountLookup(), newFakeCategoryLookup(),
		newFakeTxCreator(), fakeUserResolver{}, today)

	cases := []int64{0, -1, -100}
	for _, amt := range cases {
		req := validCreateReq(accID, catID)
		req.Amount = amt
		if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidAmount) {
			t.Errorf("amount=%d: got %v, want ErrInvalidAmount", amt, err)
		}
	}
}

func TestService_Create_RejectsInvalidFrequency(t *testing.T) {
	repo := newFakeRepo()
	accID, catID := uuid.New(), uuid.New()
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, newFakeAccountLookup(), newFakeCategoryLookup(),
		newFakeTxCreator(), fakeUserResolver{}, today)

	req := validCreateReq(accID, catID)
	req.Frequency = "hourly"
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidFrequency) {
		t.Errorf("got %v, want ErrInvalidFrequency", err)
	}
}

func TestService_Create_RejectsInvalidType(t *testing.T) {
	repo := newFakeRepo()
	accID, catID := uuid.New(), uuid.New()
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, newFakeAccountLookup(), newFakeCategoryLookup(),
		newFakeTxCreator(), fakeUserResolver{}, today)

	req := validCreateReq(accID, catID)
	req.Type = "transfer" // recurring cannot generate transfers
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrInvalidType) {
		t.Errorf("got %v, want ErrInvalidType", err)
	}
}

func TestService_Create_DefaultsIntervalToOne(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, newFakeTxCreator(), fakeUserResolver{}, today)

	req := validCreateReq(accID, catID)
	req.IntervalCount = 0
	rule, err := s.Create(uuid.New(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rule.IntervalCount != 1 {
		t.Errorf("IntervalCount = %d, want 1 (default)", rule.IntervalCount)
	}
}

func TestService_Create_DefaultsCurrencyToCOP(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, newFakeTxCreator(), fakeUserResolver{}, today)

	req := validCreateReq(accID, catID)
	req.Currency = ""
	rule, err := s.Create(uuid.New(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if rule.Currency != "COP" {
		t.Errorf("Currency = %q, want COP", rule.Currency)
	}
}

func TestService_Create_RejectsEndBeforeStart(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, newFakeTxCreator(), fakeUserResolver{}, today)

	req := validCreateReq(accID, catID)
	end := "2025-12-01"
	req.EndDate = &end
	if _, err := s.Create(uuid.New(), req); !errors.Is(err, ErrEndBeforeStart) {
		t.Errorf("got %v, want ErrEndBeforeStart", err)
	}
}

func TestService_Create_RejectsUnknownAccount(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup() // intentionally empty
	category := newFakeCategoryLookup()
	accID, catID := uuid.New(), uuid.New()
	category.addCategory(catID)
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, newFakeTxCreator(), fakeUserResolver{}, today)

	_, err := s.Create(uuid.New(), validCreateReq(accID, catID))
	if err == nil || !strings.Contains(err.Error(), "account not found") {
		t.Errorf("got %v, want account-not-found error", err)
	}
}

func TestService_Create_RejectsUnknownCategory(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup() // intentionally empty
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	today := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, newFakeTxCreator(), fakeUserResolver{}, today)

	_, err := s.Create(uuid.New(), validCreateReq(accID, catID))
	if err == nil || !strings.Contains(err.Error(), "category not found") {
		t.Errorf("got %v, want category-not-found error", err)
	}
}

func TestService_GenerateToday_NoRules(t *testing.T) {
	repo := newFakeRepo()
	today := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, newFakeAccountLookup(), newFakeCategoryLookup(),
		newFakeTxCreator(), fakeUserResolver{}, today)

	stats, err := s.GenerateToday()
	if err != nil {
		t.Fatalf("GenerateToday: %v", err)
	}
	if stats.TransactionsCreated != 0 {
		t.Errorf("TransactionsCreated = %d, want 0", stats.TransactionsCreated)
	}
	if stats.RulesScanned != 0 {
		t.Errorf("RulesScanned = %d, want 0", stats.RulesScanned)
	}
}

func TestService_GenerateToday_OneDueRule_CreatesTx(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	txC := newFakeTxCreator()
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)

	today := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, txC, fakeUserResolver{}, today)

	// Pre-seed an active rule that's due (next_run_date <= today).
	rule := &Rule{
		ID:             uuid.New(),
		UserID:         userID,
		AccountID:      accID,
		CategoryID:     catID,
		Type:           TypeExpense,
		Amount:         50000,
		Currency:       "COP",
		Frequency:      FrequencyMonthly,
		IntervalCount:  1,
		StartDate:      time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		NextRunDate:    time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		IsActive:       true,
	}
	if err := repo.CreateRule(rule); err != nil {
		t.Fatalf("seed: %v", err)
	}

	stats, err := s.GenerateToday()
	if err != nil {
		t.Fatalf("GenerateToday: %v", err)
	}
	if stats.TransactionsCreated != 1 {
		t.Errorf("TransactionsCreated = %d, want 1", stats.TransactionsCreated)
	}
	if stats.RulesScanned != 1 {
		t.Errorf("RulesScanned = %d, want 1", stats.RulesScanned)
	}
	if len(txC.calls) != 1 {
		t.Fatalf("txCreator.calls = %d, want 1", len(txC.calls))
	}
	if txC.calls[0].Amount != 50000 {
		t.Errorf("txCreator amount = %d, want 50000", txC.calls[0].Amount)
	}
	if txC.calls[0].AccountID != accID || txC.calls[0].Category != catID {
		t.Errorf("txCreator wired wrong account/category: %+v", txC.calls[0])
	}
}

func TestService_GenerateToday_TxCreatorError_RecordsFailedRun(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	txC := newFakeTxCreator()
	txC.err = errors.New("account closed")
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)

	today := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, txC, fakeUserResolver{}, today)

	rule := &Rule{
		ID:            uuid.New(),
		UserID:        userID,
		AccountID:     accID,
		CategoryID:    catID,
		Type:          TypeExpense,
		Amount:        50000,
		Currency:      "COP",
		Frequency:     FrequencyMonthly,
		IntervalCount: 1,
		StartDate:     time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		NextRunDate:   time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		IsActive:      true,
	}
	if err := repo.CreateRule(rule); err != nil {
		t.Fatalf("seed: %v", err)
	}

	stats, err := s.GenerateToday()
	if err != nil {
		t.Fatalf("GenerateToday: %v", err)
	}
	if stats.TransactionsCreated != 0 {
		t.Errorf("TransactionsCreated = %d, want 0 (tx failed)", stats.TransactionsCreated)
	}
	if len(stats.Errors) == 0 {
		t.Errorf("expected at least one error in stats")
	}
	// The run must be persisted as failed.
	runs, _ := repo.ListRunsByRule(rule.ID, 10)
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Status != RunFailed {
		t.Errorf("run status = %v, want RunFailed", runs[0].Status)
	}
	if runs[0].ErrorMessage == nil || *runs[0].ErrorMessage != "account closed" {
		t.Errorf("run error_message = %v, want 'account closed'", runs[0].ErrorMessage)
	}
}

func TestService_GenerateToday_IdempotentOnSecondRun(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	txC := newFakeTxCreator()
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)

	today := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, txC, fakeUserResolver{}, today)

	rule := &Rule{
		ID:            uuid.New(),
		UserID:        userID,
		AccountID:     accID,
		CategoryID:    catID,
		Type:          TypeExpense,
		Amount:        50000,
		Currency:      "COP",
		Frequency:     FrequencyMonthly,
		IntervalCount: 1,
		StartDate:     time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		NextRunDate:   time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC),
		IsActive:      true,
	}
	if err := repo.CreateRule(rule); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// First run: creates 1 transaction.
	if _, err := s.GenerateToday(); err != nil {
		t.Fatalf("first: %v", err)
	}
	if len(txC.calls) != 1 {
		t.Fatalf("first: txC.calls = %d, want 1", len(txC.calls))
	}

	// Second run: rule's next_run_date has been bumped to Feb, so no due rule.
	if _, err := s.GenerateToday(); err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(txC.calls) != 1 {
		t.Errorf("second: txC.calls = %d, want 1 (no new tx)", len(txC.calls))
	}
}

func TestService_RunNow_Inactive(t *testing.T) {
	repo := newFakeRepo()
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	today := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, newFakeAccountLookup(), newFakeCategoryLookup(),
		newFakeTxCreator(), fakeUserResolver{}, today)

	rule := &Rule{
		ID:            uuid.New(),
		UserID:        userID,
		AccountID:     accID,
		CategoryID:    catID,
		Type:          TypeExpense,
		Amount:        50000,
		Currency:      "COP",
		Frequency:     FrequencyMonthly,
		IntervalCount: 1,
		StartDate:     time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		NextRunDate:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		IsActive:      false,
	}
	if err := repo.CreateRule(rule); err != nil {
		t.Fatalf("seed: %v", err)
	}

	_, err := s.RunNow(rule.ID, userID)
	if err == nil || !strings.Contains(err.Error(), "not active") {
		t.Errorf("got %v, want not-active error", err)
	}
}

func TestService_RunNow_HappyPath(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	txC := newFakeTxCreator()
	userID := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)

	today := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, txC, fakeUserResolver{}, today)

	rule := &Rule{
		ID:            uuid.New(),
		UserID:        userID,
		AccountID:     accID,
		CategoryID:    catID,
		Type:          TypeIncome,
		Amount:        1500000,
		Currency:      "COP",
		Frequency:     FrequencyMonthly,
		IntervalCount: 1,
		StartDate:     time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		NextRunDate:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		IsActive:      true,
	}
	if err := repo.CreateRule(rule); err != nil {
		t.Fatalf("seed: %v", err)
	}

	txID, err := s.RunNow(rule.ID, userID)
	if err != nil {
		t.Fatalf("RunNow: %v", err)
	}
	if txID != txC.txID {
		t.Errorf("txID = %v, want %v", txID, txC.txID)
	}
	if len(txC.calls) != 1 {
		t.Fatalf("txC.calls = %d, want 1", len(txC.calls))
	}
	if txC.calls[0].Type != "income" {
		t.Errorf("type = %q, want income", txC.calls[0].Type)
	}
	// Run persisted as executed with transaction_id.
	runs, _ := repo.ListRunsByRule(rule.ID, 10)
	if len(runs) != 1 {
		t.Fatalf("runs = %d, want 1", len(runs))
	}
	if runs[0].Status != RunExecuted {
		t.Errorf("run status = %v, want RunExecuted", runs[0].Status)
	}
	if runs[0].TransactionID == nil || *runs[0].TransactionID != txID {
		t.Errorf("run transaction_id = %v, want %v", runs[0].TransactionID, txID)
	}
	// NextRunDate must have been bumped to the first strictly-after-today
	// occurrence from the new last_run_date. Today is Jan 15; monthly from
	// Jan 15 → Feb 15, but the rule's StartDate is Dec 1 2025 so its
	// cadence advances from Jan 15 by 1 month → Feb 1.
	updated, _ := repo.GetRuleByID(rule.ID, userID)
	if !updated.NextRunDate.Equal(time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("NextRunDate = %v, want 2026-02-01", updated.NextRunDate)
	}
}

func TestService_Delete_OwnershipEnforced(t *testing.T) {
	repo := newFakeRepo()
	account := newFakeAccountLookup()
	category := newFakeCategoryLookup()
	today := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	s := newTestService(repo, account, category, newFakeTxCreator(), fakeUserResolver{}, today)

	owner := uuid.New()
	other := uuid.New()
	accID, catID := uuid.New(), uuid.New()
	account.addAccount(accID)
	category.addCategory(catID)
	rule, err := s.Create(owner, validCreateReq(accID, catID))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Other user cannot delete.
	if err := s.Delete(rule.ID, other); !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("other delete: got %v, want ErrRuleNotFound", err)
	}

	// Owner can.
	if err := s.Delete(rule.ID, owner); err != nil {
		t.Errorf("owner delete: %v", err)
	}
	if _, err := repo.GetRuleByID(rule.ID, owner); !errors.Is(err, ErrRuleNotFound) {
		t.Errorf("post-delete GetRule: got %v, want ErrRuleNotFound", err)
	}
}