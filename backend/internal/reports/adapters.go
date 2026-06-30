package reports

import "github.com/google/uuid"

// CategoriesAdapter wraps a List-style function as the CategoriesLookup
// interface. The function takes (userID, categoryType) and returns the
// lightweight CategoryLite projection. Used in cmd/api/main.go to bridge
// the full categories service List to reports without exporting the
// whole service contract.
type CategoriesAdapter func(userID uuid.UUID, categoryType string) ([]CategoryLite, error)

func (a CategoriesAdapter) List(userID uuid.UUID, categoryType string) ([]CategoryLite, error) {
	if a == nil {
		return nil, nil
	}
	return a(userID, categoryType)
}

// AccountsAdapter mirrors CategoriesAdapter for accounts.
type AccountsAdapter func(userID uuid.UUID) ([]AccountLite, error)

func (a AccountsAdapter) List(userID uuid.UUID) ([]AccountLite, error) {
	if a == nil {
		return nil, nil
	}
	return a(userID)
}
