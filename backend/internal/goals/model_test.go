package goals

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPercentComplete_ZeroTarget(t *testing.T) {
	g := &Goal{TargetAmount: 0, CurrentAmount: 5000}
	if got := g.PercentComplete(); got != 0 {
		t.Errorf("zero target: got %d, want 0", got)
	}
}

func TestPercentComplete_NegativeTarget(t *testing.T) {
	g := &Goal{TargetAmount: -1, CurrentAmount: 5000}
	if got := g.PercentComplete(); got != 0 {
		t.Errorf("negative target: got %d, want 0", got)
	}
}

func TestPercentComplete_FiftyPercent(t *testing.T) {
	g := &Goal{TargetAmount: 10000, CurrentAmount: 5000}
	if got := g.PercentComplete(); got != 50 {
		t.Errorf("5000/10000: got %d, want 50", got)
	}
}

func TestPercentComplete_ClampsTo100(t *testing.T) {
	g := &Goal{TargetAmount: 1000, CurrentAmount: 1500}
	if got := g.PercentComplete(); got != 100 {
		t.Errorf("over target: got %d, want 100", got)
	}
}

func TestPercentComplete_ClampsToZero(t *testing.T) {
	// Negative current would only happen if someone manually edited the row.
	// Service-level Withdraw prevents it, but PercentComplete must defend.
	g := &Goal{TargetAmount: 1000, CurrentAmount: -100}
	if got := g.PercentComplete(); got != 0 {
		t.Errorf("negative current: got %d, want 0", got)
	}
}

func TestPercentComplete_IntegerDivision(t *testing.T) {
	// 333/1000 = 33.3% → integer division → 33.
	g := &Goal{TargetAmount: 1000, CurrentAmount: 333}
	if got := g.PercentComplete(); got != 33 {
		t.Errorf("integer truncation: got %d, want 33", got)
	}
}

func TestToDTO_NilReceiver(t *testing.T) {
	var g *Goal
	if d := g.ToDTO(); d != nil {
		t.Errorf("nil receiver: got %+v, want nil", d)
	}
}

func TestToDTO_PercentAndOverdue(t *testing.T) {
	// Use deadlines relative to "now" so the test stays correct regardless
	// of when it runs. ToDTO uses time.Now() directly.
	now := time.Now()
	past := now.AddDate(-1, 0, 0)
	future := now.AddDate(1, 0, 0)
	completedTime := past.AddDate(0, -1, 0)

	cases := []struct {
		name        string
		goal        Goal
		wantPercent int
		wantOverdue bool
	}{
		{
			name: "in-progress, no deadline",
			goal: Goal{
				ID:            uuid.New(),
				TargetAmount:  10000,
				CurrentAmount: 2500,
				IsCompleted:   false,
			},
			wantPercent: 25,
			wantOverdue: false,
		},
		{
			name: "in-progress, deadline past",
			goal: Goal{
				ID:            uuid.New(),
				TargetAmount:  10000,
				CurrentAmount: 1000,
				IsCompleted:   false,
				Deadline:      &past,
			},
			wantPercent: 10,
			wantOverdue: true,
		},
		{
			name: "in-progress, deadline future",
			goal: Goal{
				ID:            uuid.New(),
				TargetAmount:  10000,
				CurrentAmount: 1000,
				IsCompleted:   false,
				Deadline:      &future,
			},
			wantPercent: 10,
			wantOverdue: false,
		},
		{
			name: "completed, deadline past (not overdue)",
			goal: Goal{
				ID:            uuid.New(),
				TargetAmount:  10000,
				CurrentAmount: 10000,
				IsCompleted:   true,
				CompletedAt:   &completedTime,
				Deadline:      &past,
			},
			wantPercent: 100,
			wantOverdue: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			d := c.goal.ToDTO()
			if d.Percent != c.wantPercent {
				t.Errorf("Percent = %d, want %d", d.Percent, c.wantPercent)
			}
			if d.IsOverdue != c.wantOverdue {
				t.Errorf("IsOverdue = %v, want %v", d.IsOverdue, c.wantOverdue)
			}
		})
	}
}

func ptrTime(t time.Time) *time.Time { return &t }