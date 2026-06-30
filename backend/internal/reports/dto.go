package reports

import (
	"time"

	"github.com/google/uuid"
)

type RangeQuery struct {
	From string `form:"from" binding:"omitempty"`
	To   string `form:"to" binding:"omitempty"`
}

// Resolved returns the parsed [from, to] range. Missing inputs fall back to
// the current month for `from` and `time.Now()` for `to`, which matches the
// common "monthly dashboard" use case.
func (q RangeQuery) Resolved() (time.Time, time.Time, error) {
	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := now
	if q.From != "" {
		t, err := time.Parse("2006-01-02", q.From)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		from = t
	}
	if q.To != "" {
		t, err := time.Parse("2006-01-02", q.To)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		to = t
	}
	return from, to, nil
}

type ByCategoryResponse struct {
	From       time.Time             `json:"from"`
	To         time.Time             `json:"to"`
	Categories []CategoryReportItem `json:"categories"`
}

type CategoryReportItem struct {
	CategoryID uuid.UUID `json:"category_id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	Amount     int64     `json:"amount"`
	Percent    float64   `json:"percent"`
	Count      int64     `json:"count"`
}

type ByAccountResponse struct {
	From      time.Time            `json:"from"`
	To        time.Time            `json:"to"`
	Accounts  []AccountReportItem `json:"accounts"`
}

type AccountReportItem struct {
	AccountID uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
	Balance   int64     `json:"balance"`
	Income    int64     `json:"income"`
	Expense   int64     `json:"expense"`
}

type MonthlyTrendResponse struct {
	From   time.Time          `json:"from"`
	To     time.Time          `json:"to"`
	Months []MonthlyTrendItem `json:"months"`
}

type MonthlyTrendItem struct {
	Year    int   `json:"year"`
	Month   int   `json:"month"`
	Income  int64 `json:"income"`
	Expense int64 `json:"expense"`
	Net     int64 `json:"net"`
}

type BudgetVsActualResponse struct {
	From time.Time         `json:"from"`
	To   time.Time         `json:"to"`
	Rows []BudgetActualRow `json:"rows"`
}

// Summary is the dashboard's "this month vs last month" payload.
// Shape mirrors web/src/lib/schemas/report.ts → SummaryReportSchema.
type SummaryReport struct {
	From        time.Time     `json:"from"`
	To          time.Time     `json:"to"`
	TotalIncome int64          `json:"total_income"`
	TotalExpense int64         `json:"total_expense"`
	Net         int64          `json:"net"`
	ByDay       []DailySummary `json:"by_day"`
}

type DailySummary struct {
	Date   string `json:"date"`
	Income int64  `json:"income"`
	Expense int64 `json:"expense"`
}

// Cashflow mirrors web → CashflowReportSchema.
type CashflowReport struct {
	From         time.Time `json:"from"`
	To           time.Time `json:"to"`
	Income       int64     `json:"income"`
	Expense      int64     `json:"expense"`
	SavingsRate  float64   `json:"savings_rate"`
	SavingsTotal int64     `json:"savings_total"`
}