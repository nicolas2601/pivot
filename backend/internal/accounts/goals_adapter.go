package accounts

import (
	"errors"

	"github.com/google/uuid"
)

// GoalsAccountAdapter exposes the subset of accounts.Service that the
// goals package needs: "does this account exist for this user?". Keeping
// this in accounts (not goals) keeps the dependency arrow one-way.
type GoalsAccountAdapter struct {
	svc Service
}

func NewGoalsAccountAdapter(svc Service) *GoalsAccountAdapter {
	if svc == nil {
		return nil
	}
	return &GoalsAccountAdapter{svc: svc}
}

// Exists implements goals.AccountLookup. A "not found" result maps to
// (false, nil) so the goals service can distinguish "doesn't exist" from
// real DB errors (which propagate).
func (a *GoalsAccountAdapter) Exists(id, userID uuid.UUID) (bool, error) {
	_, err := a.svc.Get(id, userID)
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
