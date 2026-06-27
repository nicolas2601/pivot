package reports

import (
	"time"
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
	From      time.Time       `json:"from"`
	To        time.Time       `json:"to"`
	Breakdown []CategoryTotal `json:"breakdown"`
}

type ByAccountResponse struct {
	From      time.Time      `json:"from"`
	To        time.Time      `json:"to"`
	Breakdown []AccountTotal `json:"breakdown"`
}

type MonthlyTrendResponse struct {
	From   time.Time      `json:"from"`
	To     time.Time      `json:"to"`
	Points []MonthlyPoint `json:"points"`
}

type BudgetVsActualResponse struct {
	From time.Time         `json:"from"`
	To   time.Time         `json:"to"`
	Rows []BudgetActualRow `json:"rows"`
}