package reports

import (
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/nicolas/finanzas/backend/internal/transactions"
)

// Service re-uses the transactions.Repository to produce aggregated views
// (time series, by-category, by-account, monthly trend, budget vs actual).
//
// We deliberately keep this package thin: it owns no tables and no domain
// types — it composes queries and shapes the result for the API.
type Service struct {
	tx         transactions.Repository
	budgets    BudgetLookup
	categories CategoriesLookup
	accounts   AccountsLookup
}

type BudgetLookup interface {
	ListByUser(userID uuid.UUID) ([]BudgetSummary, error)
}

// CategoriesLookup is the minimal interface the reports layer needs to
// enrich by-category aggregations with category names and tints. Returning
// a small projection (instead of importing the categories package)
// keeps reports decoupled.
type CategoriesLookup interface {
	List(userID uuid.UUID, categoryType string) ([]CategoryLite, error)
}

// AccountsLookup mirrors CategoriesLookup for by-account aggregations.
type AccountsLookup interface {
	List(userID uuid.UUID) ([]AccountLite, error)
}

type CategoryLite struct {
	ID    uuid.UUID
	Name  string
	Color string
}

type AccountLite struct {
	ID   uuid.UUID
	Name string
}

// BudgetSummary is the minimal projection of a budget the reports layer
// needs. Returning a struct (instead of importing the budgets package)
// avoids a cyclic dependency between reports and budgets.
type BudgetSummary struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	CategoryID uuid.UUID
	Amount     int64
	Period     string
	StartDate  time.Time
	EndDate    *time.Time
}

func NewService(tx transactions.Repository, budgets BudgetLookup, cats CategoriesLookup, accs AccountsLookup) *Service {
	return &Service{tx: tx, budgets: budgets, categories: cats, accounts: accs}
}

// CategoryTotal is one row in the "expenses by category" breakdown.
type CategoryTotal struct {
	CategoryID *uuid.UUID `json:"category_id"`
	Total      int64      `json:"total"`
}

// AccountTotal is one row in the "expenses by account" breakdown.
type AccountTotal struct {
	AccountID uuid.UUID `json:"account_id"`
	Total     int64     `json:"total"`
}

// MonthlyPoint is one (year, month, total) sample in the trend line.
type MonthlyPoint struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Total int64 `json:"total"`
}

// BudgetActualRow joins a budget with its actual spending for the period.
type BudgetActualRow struct {
	BudgetID     uuid.UUID  `json:"budget_id"`
	CategoryID   uuid.UUID  `json:"category_id"`
	BudgetAmount int64      `json:"budget_amount"`
	ActualAmount int64      `json:"actual_amount"`
	Difference   int64      `json:"difference"` // actual - budget (positive = over)
	Period       string     `json:"period"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
}

// --- Queries ---

func (s *Service) ByCategory(userID uuid.UUID, from, to time.Time) ([]CategoryReportItem, error) {
	raw, err := s.tx.SumByCategory(userID, from, to)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return []CategoryReportItem{}, nil
	}
	// Lookup category names + colors in one query, build a map for O(1)
	// joins below.
	cats, err := s.categories.List(userID, "")
	if err != nil {
		return nil, err
	}
	catByID := make(map[uuid.UUID]CategoryLite, len(cats))
	for _, c := range cats {
		catByID[c.ID] = c
	}
	var total int64
	for _, r := range raw {
		total += r.Total
	}
	out := make([]CategoryReportItem, 0, len(raw))
	for _, r := range raw {
		if r.CategoryID == nil {
			continue
		}
		var pct float64
		if total > 0 {
			pct = float64(r.Total) * 100 / float64(total)
		}
		c := catByID[*r.CategoryID]
		out = append(out, CategoryReportItem{
			CategoryID: *r.CategoryID,
			Name:       c.Name,
			Color:      c.Color,
			Amount:     r.Total,
			Percent:    pct,
			Count:      0, // populated via /reports/by-category?with_count=true later if needed
		})
	}
	return out, nil
}

func (s *Service) ByAccount(userID uuid.UUID, from, to time.Time) ([]AccountReportItem, error) {
	raw, err := s.tx.SumByAccount(userID, from, to)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return []AccountReportItem{}, nil
	}
	accs, err := s.accounts.List(userID)
	if err != nil {
		return nil, err
	}
	accByID := make(map[uuid.UUID]AccountLite, len(accs))
	for _, a := range accs {
		accByID[a.ID] = a
	}
	out := make([]AccountReportItem, 0, len(raw))
	for _, r := range raw {
		a := accByID[r.AccountID]
		out = append(out, AccountReportItem{
			AccountID: r.AccountID,
			Name:      a.Name,
			Balance:   0, // Filled by /accounts/with-balance endpoint in v2.
			Income:    0,
			Expense:   r.Total,
		})
	}
	return out, nil
}

func (s *Service) MonthlyTrend(userID uuid.UUID, from, to time.Time) ([]MonthlyTrendItem, error) {
	// Returns income AND expense per month in two round-trips (one per
	// side) so the front-end can render a dual-line sparkline. In v2
	// this will be a single CTE; for now two queries is fine and still
	// O(2) instead of O(2N).
	expenseRaw, err := s.tx.MonthlyTrend(userID, from, to)
	if err != nil {
		return nil, err
	}
	incomeByMonth, err := s.monthlyIncome(userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]MonthlyTrendItem, 0, len(expenseRaw))
	for _, r := range expenseRaw {
		key := monthKey(r.Year, r.Month)
		out = append(out, MonthlyTrendItem{
			Year:    r.Year,
			Month:   r.Month,
			Income:  incomeByMonth[key],
			Expense: r.Total,
			Net:     incomeByMonth[key] - r.Total,
		})
	}
	return out, nil
}

// monthKey returns "YYYY-MM" — used as a map key for joining
// income/expense series in memory.
func monthKey(y, m int) string {
	return fmt.Sprintf("%04d-%02d", y, m)
}

// monthlyIncome groups income transactions by year-month within the
// given window. Returns a "YYYY-MM" → total map for easy join with the
// expense series in MonthlyTrend.
func (s *Service) monthlyIncome(userID uuid.UUID, from, to time.Time) (map[string]int64, error) {
	rows, err := s.tx.MonthlyTrendAmountsByMonth(userID, from, to, string(transactions.TypeIncome))
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[monthKey(r.Year, r.Month)] = r.Total
	}
	return out, nil
}

// Summary computes the dashboard's headline numbers plus a per-day
// breakdown. Two aggregate queries (income by day, expense by day), joined
// in memory.
func (s *Service) Summary(userID uuid.UUID, from, to time.Time) (*SummaryReport, error) {
	incByDay, err := s.tx.AmountsByDay(userID, from, to, string(transactions.TypeIncome))
	if err != nil {
		return nil, err
	}
	expByDay, err := s.tx.AmountsByDay(userID, from, to, string(transactions.TypeExpense))
	if err != nil {
		return nil, err
	}
	days := make(map[string]bool, len(incByDay)+len(expByDay))
	for d := range incByDay {
		days[d] = true
	}
	for d := range expByDay {
		days[d] = true
	}
	type kv struct {
		date string
		d    time.Time
	}
	pairs := make([]kv, 0, len(days))
	for d := range days {
		t, err := time.Parse("2006-01-02", d)
		if err != nil {
			continue
		}
		pairs = append(pairs, kv{date: d, d: t})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].d.Before(pairs[j].d) })
	byDay := make([]DailySummary, 0, len(pairs))
	var totalIncome, totalExpense int64
	for _, p := range pairs {
		inc := incByDay[p.date]
		exp := expByDay[p.date]
		byDay = append(byDay, DailySummary{Date: p.date, Income: inc, Expense: exp})
		totalIncome += inc
		totalExpense += exp
	}
	return &SummaryReport{
		From:        from,
		To:          to,
		TotalIncome: totalIncome,
		TotalExpense: totalExpense,
		Net:         totalIncome - totalExpense,
		ByDay:       byDay,
	}, nil
}

// Cashflow returns aggregate income/expense for the window plus a
// savings rate.
func (s *Service) Cashflow(userID uuid.UUID, from, to time.Time) (*CashflowReport, error) {
	incByDay, err := s.tx.AmountsByDay(userID, from, to, string(transactions.TypeIncome))
	if err != nil {
		return nil, err
	}
	expByDay, err := s.tx.AmountsByDay(userID, from, to, string(transactions.TypeExpense))
	if err != nil {
		return nil, err
	}
	var income, expense int64
	for _, v := range incByDay {
		income += v
	}
	for _, v := range expByDay {
		expense += v
	}
	var rate float64
	if income+expense > 0 {
		rate = float64(income-expense) * 100 / float64(income+expense)
	}
	return &CashflowReport{
		From:         from,
		To:           to,
		Income:       income,
		Expense:      expense,
		SavingsRate:  rate,
		SavingsTotal: income - expense,
	}, nil
}

// BudgetVsActual computes each budget's actual spending for the period that
// overlaps [from, to].
//
// Two-pass approach:
//   1. Round up the budgets into a single UNION-style date range — the
//      union of all budgets' date windows is at most [min(StartDate),
//      max(EndDate or `to`)], and that fits within the user's request
//      window which is itself a superset.
//   2. Issue ONE aggregate query that returns per-category totals for
//      the entire union.
//   3. For each budget, take the value for its category from the cache
//      when that budget's [StartDate, EndDate or `to`] is fully contained
//      in the union, otherwise fall back to a per-budget aggregated query.
//
// Per-budget windows outside the unified cache get a single per-budget
// query (the migration 000010 covering index makes these O(log n)).
func (s *Service) BudgetVsActual(userID uuid.UUID, from, to time.Time) ([]BudgetActualRow, error) {
	if s.budgets == nil {
		return nil, nil
	}
	budgets, err := s.budgets.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	if len(budgets) == 0 {
		return []BudgetActualRow{}, nil
	}
	// Compute the cache range: extend the user's window to cover any
	// budget whose window starts before `from` or ends after `to`. This way
	// every per-budget filter is a sub-range of the cache and we never
	// miss data, while still bounding the cache to the budgets' lifetimes.
	cacheFrom := from
	cacheTo := to
	for _, b := range budgets {
		if b.StartDate.Before(cacheFrom) {
			cacheFrom = b.StartDate
		}
		if b.EndDate != nil && b.EndDate.After(cacheTo) {
			cacheTo = *b.EndDate
		}
	}
	cache, err := s.tx.SumByCategory(userID, cacheFrom, cacheTo)
	if err != nil {
		return nil, err
	}
	actualsByCategory := make(map[uuid.UUID]int64, len(cache))
	for _, a := range cache {
		if a.CategoryID == nil {
			continue
		}
		actualsByCategory[*a.CategoryID] = a.Total
	}
	rows := make([]BudgetActualRow, 0, len(budgets))
	for _, b := range budgets {
		bFrom := b.StartDate
		if bFrom.Before(from) {
			bFrom = from
		}
		bTo := to
		if b.EndDate != nil && b.EndDate.Before(to) {
			bTo = *b.EndDate
		}
		// If the budget's [bFrom, bTo] is fully contained in the cache
		// range, use the cached aggregate. Otherwise fall back to a single
		// per-budget query — guarded so we don't loop over budgets here.
		var actual int64
		if !bFrom.Before(cacheFrom) && !bTo.After(cacheTo) {
			actual = actualsByCategory[b.CategoryID]
		} else {
			agg, err := s.tx.SumByCategory(userID, bFrom, bTo)
			if err != nil {
				return nil, err
			}
			for _, a := range agg {
				if a.CategoryID != nil && *a.CategoryID == b.CategoryID {
					actual = a.Total
					break
				}
			}
		}
		rows = append(rows, BudgetActualRow{
			BudgetID:     b.ID,
			CategoryID:   b.CategoryID,
			BudgetAmount: b.Amount,
			ActualAmount: actual,
			Difference:   actual - b.Amount,
			Period:       b.Period,
			StartDate:    b.StartDate,
			EndDate:      b.EndDate,
		})
	}
	return rows, nil
}