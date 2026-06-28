package goals

import "time"

// GoalDTO is the wire shape — Goal with computed fields the UI needs but
// doesn't need to mutate (percent, is_overdue).
type GoalDTO struct {
	*Goal
	Percent     int  `json:"percent"`
	IsOverdue   bool `json:"is_overdue"`
}

// ToDTO returns the public shape with derived fields. Safe on nil receiver
// for handler convenience.
func (g *Goal) ToDTO() *GoalDTO {
	if g == nil {
		return nil
	}
	overdue := false
	if g.Deadline != nil && !g.IsCompleted && g.Deadline.Before(time.Now()) {
		overdue = true
	}
	return &GoalDTO{
		Goal:       g,
		Percent:    g.PercentComplete(),
		IsOverdue:  overdue,
	}
}
