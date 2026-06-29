package recurring

import (
	"testing"
	"time"
)

func TestIsValidFrequency(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"daily", true},
		{"weekly", true},
		{"biweekly", true},
		{"monthly", true},
		{"yearly", true},
		{"hourly", false},
		{"biennial", false},
		{"", false},
		{"DAILY", false}, // case-sensitive on purpose; service layer normalizes
	}
	for _, c := range cases {
		got := IsValidFrequency(c.in)
		if got != c.want {
			t.Errorf("IsValidFrequency(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestIsValidTxType(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"expense", true},
		{"income", true},
		{"transfer", false}, // recurring cannot generate transfers
		{"", false},
		{"EXPENSE", false},
	}
	for _, c := range cases {
		got := IsValidTxType(c.in)
		if got != c.want {
			t.Errorf("IsValidTxType(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// Helper to build a Rule at a known start date with a fixed interval.
func newRule(freq Frequency, interval int, start time.Time) *Rule {
	return &Rule{
		Frequency:     freq,
		IntervalCount: interval,
		StartDate:     start,
	}
}

func TestNextOccurrence_Daily(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyDaily, 1, start)

	// Strictly after start.
	got := r.NextOccurrence(start)
	want := start.AddDate(0, 0, 1)
	if got == nil || !got.Equal(want) {
		t.Errorf("NextOccurrence(start) = %v, want %v", got, want)
	}

	// Two days later.
	from := start.AddDate(0, 0, 2)
	got = r.NextOccurrence(from)
	want = start.AddDate(0, 0, 3)
	if got == nil || !got.Equal(want) {
		t.Errorf("NextOccurrence(+2d) = %v, want %v", got, want)
	}
}

func TestNextOccurrence_WeeklyAndBiweekly(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	weekly := newRule(FrequencyWeekly, 1, start)
	if got := weekly.NextOccurrence(start); got == nil || !got.Equal(start.AddDate(0, 0, 7)) {
		t.Errorf("weekly: got %v, want %v", got, start.AddDate(0, 0, 7))
	}

	biweekly := newRule(FrequencyBiweekly, 1, start)
	if got := biweekly.NextOccurrence(start); got == nil || !got.Equal(start.AddDate(0, 0, 14)) {
		t.Errorf("biweekly: got %v, want %v", got, start.AddDate(0, 0, 14))
	}
}

func TestNextOccurrence_Monthly(t *testing.T) {
	start := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyMonthly, 1, start)
	got := r.NextOccurrence(start)
	want := time.Date(2026, 2, 15, 0, 0, 0, 0, time.UTC)
	if got == nil || !got.Equal(want) {
		t.Errorf("monthly: got %v, want %v", got, want)
	}

	// Year-rollover: start mid-month, no edge case.
	startDec := time.Date(2026, 12, 15, 0, 0, 0, 0, time.UTC)
	r = newRule(FrequencyMonthly, 1, startDec)
	got = r.NextOccurrence(startDec)
	want = time.Date(2027, 1, 15, 0, 0, 0, 0, time.UTC)
	if got == nil || !got.Equal(want) {
		t.Errorf("monthly year-rollover: got %v, want %v", got, want)
	}
}

// TestNextOccurrence_MonthlyEndOfMonth documents the current behavior when
// the start day doesn't exist in the next month (Go time.AddDate normalizes
// the overflow to the next-next month). If you want "last day of month"
// semantics for end-of-month billing, that needs a different implementation.
func TestNextOccurrence_MonthlyEndOfMonth(t *testing.T) {
	start := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyMonthly, 1, start)
	got := r.NextOccurrence(start)
	// Go normalizes Jan 31 + 1 month = Mar 3 (Feb 31 → Mar 3).
	want := time.Date(2026, 3, 3, 0, 0, 0, 0, time.UTC)
	if got == nil || !got.Equal(want) {
		t.Errorf("monthly jan31+1m: got %v, want %v (Go AddDate overflow)", got, want)
	}
}

func TestNextOccurrence_Yearly(t *testing.T) {
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyYearly, 1, start)
	got := r.NextOccurrence(start)
	want := time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)
	if got == nil || !got.Equal(want) {
		t.Errorf("yearly: got %v, want %v", got, want)
	}
}

func TestNextOccurrence_IntervalGreaterThanOne(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// Every 3 days.
	r := newRule(FrequencyDaily, 3, start)
	got := r.NextOccurrence(start)
	want := start.AddDate(0, 0, 3)
	if got == nil || !got.Equal(want) {
		t.Errorf("daily every-3: got %v, want %v", got, want)
	}

	// Every 2 months.
	r = newRule(FrequencyMonthly, 2, start)
	got = r.NextOccurrence(start)
	want = time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if got == nil || !got.Equal(want) {
		t.Errorf("monthly every-2: got %v, want %v", got, want)
	}
}

func TestNextOccurrence_PastEndDate(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 8, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyDaily, 1, start)
	r.EndDate = &end

	// After end → nil.
	got := r.NextOccurrence(end.AddDate(0, 0, 1))
	if got != nil {
		t.Errorf("after end_date: got %v, want nil", got)
	}
}

func TestOccurrencesBetween(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyDaily, 1, start)

	// 5 days: Jan 2, 3, 4, 5, 6.
	got := r.OccurrencesBetween(start, time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC))
	want := []time.Time{
		start.AddDate(0, 0, 1),
		start.AddDate(0, 0, 2),
		start.AddDate(0, 0, 3),
		start.AddDate(0, 0, 4),
		start.AddDate(0, 0, 5),
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got=%v", len(got), len(want), got)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("[%d] got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestOccurrencesBetween_EmptyWhenFromAfterTo(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyDaily, 1, start)
	got := r.OccurrencesBetween(
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestOccurrencesBetween_RespectsEndDate(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 4, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyDaily, 1, start)
	r.EndDate = &end

	// Query 30 days but EndDate stops at Jan 4.
	got := r.OccurrencesBetween(start, time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC))
	want := []time.Time{
		start.AddDate(0, 0, 1),
		start.AddDate(0, 0, 2),
		start.AddDate(0, 0, 3),
	}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d; got=%v", len(got), len(want), got)
	}
	for i := range want {
		if !got[i].Equal(want[i]) {
			t.Errorf("[%d] got %v, want %v", i, got[i], want[i])
		}
	}
}

func TestLastRunDateOrZero(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	r := newRule(FrequencyDaily, 1, start)

	// No last_run_date → returns start - 1 day, ensuring NextOccurrence picks start.
	if got := r.LastRunDateOrZero(); !got.Equal(start.AddDate(0, 0, -1)) {
		t.Errorf("zero: got %v, want %v", got, start.AddDate(0, 0, -1))
	}

	// With last_run_date → returns it.
	last := time.Date(2026, 1, 5, 0, 0, 0, 0, time.UTC)
	r.LastRunDate = &last
	if got := r.LastRunDateOrZero(); !got.Equal(last) {
		t.Errorf("with last: got %v, want %v", got, last)
	}
}