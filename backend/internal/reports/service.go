package reports

import (
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
	tx      transactions.Repository
	budgets BudgetLookup
}

type BudgetLookup interface {
	ListByUser(userID uuid.UUID) ([]BudgetSummary, error)
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

func NewService(tx transactions.Repository, budgets BudgetLookup) *Service {
	return &Service{tx: tx, budgets: budgets}
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

func (s *Service) ByCategory(userID uuid.UUID, from, to time.Time) ([]CategoryTotal, error) {
	raw, err := s.tx.SumByCategory(userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]CategoryTotal, 0, len(raw))
	for _, r := range raw {
		out = append(out, CategoryTotal{CategoryID: r.CategoryID, Total: r.Total})
	}
	return out, nil
}

func (s *Service) ByAccount(userID uuid.UUID, from, to time.Time) ([]AccountTotal, error) {
	raw, err := s.tx.SumByAccount(userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]AccountTotal, 0, len(raw))
	for _, r := range raw {
		out = append(out, AccountTotal{AccountID: r.AccountID, Total: r.Total})
	}
	return out, nil
}

func (s *Service) MonthlyTrend(userID uuid.UUID, from, to time.Time) ([]MonthlyPoint, error) {
	raw, err := s.tx.MonthlyTrend(userID, from, to)
	if err != nil {
		return nil, err
	}
	out := make([]MonthlyPoint, 0, len(raw))
	for _, r := range raw {
		out = append(out, MonthlyPoint{Year: r.Year, Month: r.Month, Total: r.Total})
	}
	return out, nil
}

// BudgetVsActual computes each budget's actual spending for the period that
// overlaps [from, to]. The aggregation is approximate — we sum every
// expense in the budget's window — which matches the common UX expectation
// of "where am I vs my limit this month".
func (s *Service) BudgetVsActual(userID uuid.UUID, from, to time.Time) ([]BudgetActualRow, error) {
	if s.budgets == nil {
		return nil, nil
	}
	budgets, err := s.budgets.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	rows := make([]BudgetActualRow, 0, len(budgets))
	for _, b := range budgets {
		// Translate budgets.Date range into [from, to]. If a budget has no
		// EndDate we use `to` (assume open-ended).
		bFrom := b.StartDate
		if bFrom.Before(from) {
			bFrom = from
		}
		bTo := to
		if b.EndDate != nil && b.EndDate.Before(to) {
			bTo = *b.EndDate
		}
		// We need spending filtered by category AND window. SumByCategory
		// doesn't filter by category so we filter manually here.
		actuals, err := s.tx.SumByCategory(userID, bFrom, bTo)
		if err != nil {
			return nil, err
		}
		var actual int64
		for _, a := range actuals {
			if a.CategoryID == nil {
				continue
			}
			if *a.CategoryID == b.CategoryID {
				actual = a.Total
				break
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