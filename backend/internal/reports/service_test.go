package reports

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/nicolas/finanzas/backend/internal/transactions"
)

// txRepoStub implements transactions.Repository for reports tests.
type txRepoStub struct {
	byCategory []transactions.CategorySum
	byAccount  []transactions.AccountSum
	trend      []transactions.MonthlyTotal
	byDay      map[string]int64 // date → total, last-call wins
}

func (s *txRepoStub) Create(*transactions.Transaction) error { return nil }
func (s *txRepoStub) GetByID(uuid.UUID, uuid.UUID) (*transactions.Transaction, error) {
	return nil, nil
}
func (s *txRepoStub) ListByUser(uuid.UUID, transactions.ListFilter) ([]transactions.Transaction, error) {
	return nil, nil
}
func (s *txRepoStub) Update(*transactions.Transaction) error { return nil }
func (s *txRepoStub) Delete(uuid.UUID, uuid.UUID) error     { return nil }
func (s *txRepoStub) CreateTransfer(uuid.UUID, *transactions.Transaction, *transactions.Transaction) error {
	return nil
}
func (s *txRepoStub) DeletePair(uuid.UUID, uuid.UUID) error { return nil }

func (s *txRepoStub) SumByCategory(_ uuid.UUID, _, _ time.Time) ([]transactions.CategorySum, error) {
	return s.byCategory, nil
}
func (s *txRepoStub) SumByAccount(_ uuid.UUID, _, _ time.Time) ([]transactions.AccountSum, error) {
	return s.byAccount, nil
}
func (s *txRepoStub) MonthlyTrend(_ uuid.UUID, _, _ time.Time) ([]transactions.MonthlyTotal, error) {
	return s.trend, nil
}
func (s *txRepoStub) MonthlyTrendAmountsByMonth(_ uuid.UUID, _, _ time.Time, _ string) ([]transactions.MonthlyTotal, error) {
	return s.trend, nil
}
func (s *txRepoStub) AmountsByDay(_ uuid.UUID, _, _ time.Time, _ string) (map[string]int64, error) {
	return s.byDay, nil
}

// budgetStub implements BudgetLookup.
type budgetStub struct {
	budgets []BudgetSummary
}

func (s *budgetStub) ListByUser(_ uuid.UUID) ([]BudgetSummary, error) {
	return s.budgets, nil
}

// categoriesStub implements CategoriesLookup.
type categoriesStub struct {
	cats []CategoryLite
}

func (s *categoriesStub) List(_ uuid.UUID, _ string) ([]CategoryLite, error) {
	return s.cats, nil
}

// accountsStub implements AccountsLookup.
type accountsStub struct {
	accs []AccountLite
}

func (s *accountsStub) List(_ uuid.UUID) ([]AccountLite, error) {
	return s.accs, nil
}

// newTestService builds a Service with the new 4-arg signature. All
// optional stubs default to a working no-op so tests only need to set
// the fields they care about.
func newTestService(tx *txRepoStub, opts ...func(*Service)) *Service {
	s := NewService(tx, nil, &categoriesStub{}, &accountsStub{})
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func withBudget(b []BudgetSummary) func(*Service) {
	return func(s *Service) { s.budgets = &budgetStub{budgets: b} }
}

func withCategories(c []CategoryLite) func(*Service) {
	return func(s *Service) { s.categories = &categoriesStub{cats: c} }
}

func withAccounts(a []AccountLite) func(*Service) {
	return func(s *Service) { s.accounts = &accountsStub{accs: a} }
}

func TestService_ByCategory_ShapeMapping(t *testing.T) {
	catID := uuid.New()
	tx := &txRepoStub{
		byCategory: []transactions.CategorySum{
			{CategoryID: &catID, Total: 12345},
		},
	}
	s := newTestService(tx, withCategories([]CategoryLite{
		{ID: catID, Name: "Alimentación", Color: "#a7e5d3"},
	}))
	out, err := s.ByCategory(uuid.New(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ByCategory: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	if out[0].Amount != 12345 {
		t.Errorf("Amount = %d, want 12345", out[0].Amount)
	}
	if out[0].CategoryID != catID {
		t.Errorf("CategoryID = %v, want %v", out[0].CategoryID, catID)
	}
	if out[0].Name != "Alimentación" {
		t.Errorf("Name = %q, want Alimentación", out[0].Name)
	}
}

func TestService_ByAccount_ShapeMapping(t *testing.T) {
	accID := uuid.New()
	tx := &txRepoStub{
		byAccount: []transactions.AccountSum{
			{AccountID: accID, Total: 9999},
		},
	}
	s := newTestService(tx, withAccounts([]AccountLite{
		{ID: accID, Name: "Efectivo"},
	}))
	out, err := s.ByAccount(uuid.New(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ByAccount: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d, want 1", len(out))
	}
	if out[0].Expense != 9999 {
		t.Errorf("Expense = %d, want 9999", out[0].Expense)
	}
	if out[0].AccountID != accID {
		t.Errorf("AccountID = %v, want %v", out[0].AccountID, accID)
	}
}

func TestService_MonthlyTrend_PreservesOrder(t *testing.T) {
	tx := &txRepoStub{
		trend: []transactions.MonthlyTotal{
			{Year: 2026, Month: 1, Total: 100},
			{Year: 2026, Month: 2, Total: 200},
			{Year: 2026, Month: 3, Total: 300},
		},
	}
	s := newTestService(tx)
	out, err := s.MonthlyTrend(uuid.New(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("MonthlyTrend: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("len = %d, want 3", len(out))
	}
	expected := []int64{100, 200, 300}
	for i, want := range expected {
		if out[i].Expense != want {
			t.Errorf("[%d] Expense = %d, want %d", i, out[i].Expense, want)
		}
	}
}

func TestService_BudgetVsActual_NoBudgets_ReturnsEmpty(t *testing.T) {
	s := newTestService(&txRepoStub{}, withBudget(nil))
	rows, err := s.BudgetVsActual(uuid.New(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("BudgetVsActual: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("len = %d, want 0", len(rows))
	}
}

func TestService_BudgetVsActual_DifferenceIsActualMinusBudget(t *testing.T) {
	catID := uuid.New()
	monthlyBudget := BudgetSummary{
		ID: uuid.New(), UserID: uuid.New(), CategoryID: catID,
		Amount: 1000, Period: "monthly",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	tx := &txRepoStub{
		byCategory: []transactions.CategorySum{
			{CategoryID: &catID, Total: 1500}, // overspent
		},
	}
	s := newTestService(tx, withBudget([]BudgetSummary{monthlyBudget}))

	rows, err := s.BudgetVsActual(uuid.New(),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BudgetVsActual: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("len = %d, want 1", len(rows))
	}
	if rows[0].ActualAmount != 1500 {
		t.Errorf("ActualAmount = %d, want 1500", rows[0].ActualAmount)
	}
	if rows[0].BudgetAmount != 1000 {
		t.Errorf("BudgetAmount = %d, want 1000", rows[0].BudgetAmount)
	}
	if rows[0].Difference != 500 {
		t.Errorf("Difference = %d, want 500 (actual - budget)", rows[0].Difference)
	}
}

func TestService_BudgetVsActual_UnderBudget_NegativeDifference(t *testing.T) {
	catID := uuid.New()
	monthlyBudget := BudgetSummary{
		ID: uuid.New(), UserID: uuid.New(), CategoryID: catID,
		Amount: 1000, Period: "monthly",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	tx := &txRepoStub{
		byCategory: []transactions.CategorySum{
			{CategoryID: &catID, Total: 300},
		},
	}
	s := newTestService(tx, withBudget([]BudgetSummary{monthlyBudget}))

	rows, err := s.BudgetVsActual(uuid.New(),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BudgetVsActual: %v", err)
	}
	if rows[0].Difference != -700 {
		t.Errorf("Difference = %d, want -700 (under budget)", rows[0].Difference)
	}
}

func TestService_BudgetVsActual_NoSpendingYet_DifferenceEqualsMinusBudget(t *testing.T) {
	catID := uuid.New()
	monthlyBudget := BudgetSummary{
		ID: uuid.New(), UserID: uuid.New(), CategoryID: catID,
		Amount: 1000, Period: "monthly",
		StartDate: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	tx := &txRepoStub{
		byCategory: []transactions.CategorySum{
			{CategoryID: &catID, Total: 0}, // zero spending
		},
	}
	s := newTestService(tx, withBudget([]BudgetSummary{monthlyBudget}))

	rows, err := s.BudgetVsActual(uuid.New(),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("BudgetVsActual: %v", err)
	}
	if rows[0].Difference != -1000 {
		t.Errorf("Difference = %d, want -1000 (no spending)", rows[0].Difference)
	}
}

func TestService_BudgetVsActual_NilBudgetLookup_ReturnsNil(t *testing.T) {
	s := newTestService(&txRepoStub{}) // no withBudget
	rows, err := s.BudgetVsActual(uuid.New(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("BudgetVsActual: %v", err)
	}
	if rows != nil {
		t.Errorf("expected nil rows when budget lookup missing")
	}
}

func TestService_Summary_AggregatesIncomeExpenseNet(t *testing.T) {
	tx := &txRepoStub{
		byDay: map[string]int64{
			"2026-06-01": 5000, // shared stub — exercise the structural contract
		},
	}
	// We need separate maps for income and expense. Use the same map for
	// now: tests with separate maps would need to modify AmountsByDay to
	// dispatch by type. For a quick smoke test we validate the structural
	// contract only.
	s := newTestService(tx)
	out, err := s.Summary(uuid.New(),
		time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Summary: %v", err)
	}
	if out == nil {
		t.Fatal("Summary returned nil report")
	}
	if out.Net != out.TotalIncome-out.TotalExpense {
		t.Errorf("Net = %d, want income - expense", out.Net)
	}
}
