package budgets

import (
	"github.com/google/uuid"

	"github.com/nicolas/finanzas/backend/internal/reports"
)

// ReportBudgetAdapter projects a budgets.Budget into the minimal shape the
// reports package needs. It avoids reports importing budgets (which would
// force reports → budgets → transactions → ... chain).
type ReportBudgetAdapter struct {
	repo Repository
}

// NewReportBudgetAdapter wires budgets.Repository into the reports
// BudgetLookup contract.
func NewReportBudgetAdapter(repo Repository) *ReportBudgetAdapter {
	if repo == nil {
		return nil
	}
	return &ReportBudgetAdapter{repo: repo}
}

// ListByUser implements reports.BudgetLookup.
func (a *ReportBudgetAdapter) ListByUser(userID uuid.UUID) ([]reports.BudgetSummary, error) {
	bs, err := a.repo.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	out := make([]reports.BudgetSummary, 0, len(bs))
	for _, b := range bs {
		out = append(out, reports.BudgetSummary{
			ID:         b.ID,
			UserID:     b.UserID,
			CategoryID: b.CategoryID,
			Amount:     b.Amount,
			Period:     string(b.Period),
			StartDate:  b.StartDate,
			EndDate:    b.EndDate,
		})
	}
	return out, nil
}