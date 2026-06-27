package auth

import (
	"errors"

	"github.com/google/uuid"
)

// TravelUserAdapter lets the travel package resolve an email to a userID
// without importing the auth package directly. The travel package only
// needs the UUID, so the contract is intentionally minimal.
//
// We declare a tiny private interface (FindByEmailer) rather than importing
// travel to avoid import cycles. main.go does the structural wiring.
type TravelUserAdapter struct {
	repo UserRepository
}

// FindByEmailer is the contract travel expects from this adapter. The travel
// package owns the canonical type; we mirror the shape here so the wiring
// in main.go can assign *TravelUserAdapter directly.
type FindByEmailer interface {
	FindByEmail(email string) (uuid.UUID, error)
}

// NewTravelUserAdapter wires auth.UserRepository into the travel.UserLookup
// contract.
func NewTravelUserAdapter(repo UserRepository) *TravelUserAdapter {
	if repo == nil {
		return nil
	}
	return &TravelUserAdapter{repo: repo}
}

// FindByEmail satisfies the travel.UserLookup contract by signature.
func (a *TravelUserAdapter) FindByEmail(email string) (uuid.UUID, error) {
	user, err := a.repo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return uuid.Nil, err
		}
		return uuid.Nil, err
	}
	return user.ID, nil
}