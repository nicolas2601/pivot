package transactions

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Account is a minimal interface so the service can validate account ownership
// and currency without depending on the accounts package directly.
type Account interface {
	GetByID(id, userID uuid.UUID) (AccountInfo, error)
}

type AccountInfo struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Currency string
}

func (a AccountInfo) IsOwnedBy(userID uuid.UUID) bool {
	return a.UserID == userID
}

// CategoryLookup is the optional contract for category validation. Transfer
// transactions do not require a category, so we let nil be the "skip" signal.
type CategoryLookup interface {
	GetByID(id, userID uuid.UUID) (exists bool, err error)
}

type Service struct {
	repo      Repository
	accounts  Account
	categories CategoryLookup
}

func NewService(repo Repository, accounts Account, categories CategoryLookup) *Service {
	return &Service{repo: repo, accounts: accounts, categories: categories}
}

var (
	ErrInvalidType     = fmt.Errorf("invalid transaction type")
	ErrAccountNotFound = fmt.Errorf("account not found")
	ErrCategoryNotFound = fmt.Errorf("category not found")
	ErrInvalidAmount   = fmt.Errorf("amount must be greater than zero")
	ErrSameAccount     = ErrAccountMismatch
)

// ErrCurrencyMismatch is defined in errors.go and shared with the repository.
// Handler and tests can match it via the service namespace.

type CreateRequest struct {
	AccountID   string  `json:"account_id"`
	CategoryID  *string `json:"category_id,omitempty"`
	Type        string  `json:"type"`
	Amount      int64   `json:"amount"`
	Currency    string  `json:"currency"`
	Date        string  `json:"date"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type UpdateRequest struct {
	CategoryID  *string `json:"category_id,omitempty"`
	Amount      *int64  `json:"amount,omitempty"`
	Date        *string `json:"date,omitempty"`
	Description *string `json:"description,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

type TransferRequest struct {
	FromAccountID string  `json:"from_account_id"`
	ToAccountID   string  `json:"to_account_id"`
	Amount        int64   `json:"amount"`
	Currency      string  `json:"currency"`
	Date          string  `json:"date"`
	Description   *string `json:"description,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

type TransferResult struct {
	Source *Transaction `json:"source"`
	Dest   *Transaction `json:"dest"`
}

func (s *Service) Create(userID uuid.UUID, req CreateRequest) (*Transaction, error) {
	if !IsValidType(req.Type) {
		return nil, ErrInvalidType
	}
	if TxType(req.Type) == TypeTransfer {
		return nil, fmt.Errorf("use Transfer() to create transfer transactions")
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("parse account_id: %w", err)
	}
	account, err := s.accounts.GetByID(accountID, userID)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	var categoryID *uuid.UUID
	if req.CategoryID != nil && *req.CategoryID != "" {
		cid, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("parse category_id: %w", err)
		}
		if s.categories != nil {
			exists, err := s.categories.GetByID(cid, userID)
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, ErrCategoryNotFound
			}
		}
		categoryID = &cid
	}

	date, err := parseDate(req.Date)
	if err != nil {
		return nil, err
	}

	tx := &Transaction{
		UserID:      userID,
		AccountID:   accountID,
		CategoryID:  categoryID,
		Type:        TxType(req.Type),
		Amount:      req.Amount,
		Currency:    currencyOrDefault(req.Currency, account.Currency),
		Date:        date,
		Description: req.Description,
		Notes:       req.Notes,
	}
	if err := s.repo.Create(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *Service) Get(id, userID uuid.UUID) (*Transaction, error) {
	return s.repo.GetByID(id, userID)
}

// CreateFromRecurring is the entry point used by the recurring engine.
// Skips transfer-rejection (recurring can only produce expense/income),
// uses the provided account & category (already validated by the caller),
// and stamps RecurringRunID so the run record links back.
func (s *Service) CreateFromRecurring(
	userID, accountID, categoryID uuid.UUID,
	txType string,
	amount int64,
	currency string,
	date time.Time,
	description, notes *string,
	recurringRunID uuid.UUID,
) (uuid.UUID, error) {
	if !IsValidType(txType) || TxType(txType) == TypeTransfer {
		return uuid.Nil, ErrInvalidType
	}
	if amount <= 0 {
		return uuid.Nil, ErrInvalidAmount
	}
	account, err := s.accounts.GetByID(accountID, userID)
	if err != nil {
		return uuid.Nil, ErrAccountNotFound
	}
	if s.categories != nil {
		exists, err := s.categories.GetByID(categoryID, userID)
		if err != nil {
			return uuid.Nil, err
		}
		if !exists {
			return uuid.Nil, ErrCategoryNotFound
		}
	}
	cid := categoryID
	tx := &Transaction{
		UserID:      userID,
		AccountID:   accountID,
		CategoryID:  &cid,
		Type:        TxType(txType),
		Amount:      amount,
		Currency:    currencyOrDefault(currency, account.Currency),
		Date:        date,
		Description: description,
		Notes:       notes,
	}
	if err := s.repo.Create(tx); err != nil {
		return uuid.Nil, err
	}
	return tx.ID, nil
}

func (s *Service) List(userID uuid.UUID, f ListFilter) ([]Transaction, error) {
	return s.repo.ListByUser(userID, f)
}

func (s *Service) Update(id, userID uuid.UUID, req UpdateRequest) (*Transaction, error) {
	tx, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	if tx.Type == TypeTransfer {
		return nil, fmt.Errorf("transfers are immutable; create a reversing transfer instead")
	}
	if req.CategoryID != nil {
		if *req.CategoryID == "" {
			tx.CategoryID = nil
		} else {
			cid, err := uuid.Parse(*req.CategoryID)
			if err != nil {
				return nil, fmt.Errorf("parse category_id: %w", err)
			}
			if s.categories != nil {
				exists, err := s.categories.GetByID(cid, userID)
				if err != nil {
					return nil, err
				}
				if !exists {
					return nil, ErrCategoryNotFound
				}
			}
			tx.CategoryID = &cid
		}
	}
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, ErrInvalidAmount
		}
		tx.Amount = *req.Amount
	}
	if req.Date != nil {
		date, err := parseDate(*req.Date)
		if err != nil {
			return nil, err
		}
		tx.Date = date
	}
	if req.Description != nil {
		tx.Description = req.Description
	}
	if req.Notes != nil {
		tx.Notes = req.Notes
	}
	if err := s.repo.Update(tx); err != nil {
		return nil, err
	}
	return tx, nil
}

func (s *Service) Delete(id, userID uuid.UUID) error {
	tx, err := s.repo.GetByID(id, userID)
	if err != nil {
		return err
	}
	if tx.Type == TypeTransfer && tx.TransferPairID != nil {
		// Cascade soft-delete both legs so the pair cannot be orphaned.
		return s.repo.DeletePair(*tx.TransferPairID, userID)
	}
	return s.repo.Delete(id, userID)
}

// Transfer creates an atomic pair of transactions moving `amount` from
// `fromAccountID` to `toAccountID`. Both transactions are tagged with
// TypeTransfer and share the same TransferPairID.
//
// On the source side we record an expense-style transfer (money leaving).
// On the destination side we record an income-style transfer (money arriving).
// Both rows have the same amount so balance calculations stay consistent.
func (s *Service) Transfer(userID uuid.UUID, req TransferRequest) (*TransferResult, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
	}
	fromID, err := uuid.Parse(req.FromAccountID)
	if err != nil {
		return nil, fmt.Errorf("parse from_account_id: %w", err)
	}
	toID, err := uuid.Parse(req.ToAccountID)
	if err != nil {
		return nil, fmt.Errorf("parse to_account_id: %w", err)
	}
	from, err := s.accounts.GetByID(fromID, userID)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	to, err := s.accounts.GetByID(toID, userID)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	if from.Currency != to.Currency {
		return nil, ErrCurrencyMismatch
	}

	date, err := parseDate(req.Date)
	if err != nil {
		return nil, err
	}

	currency := req.Currency
	if currency == "" {
		currency = from.Currency
	}

	source := &Transaction{
		UserID:      userID,
		AccountID:   fromID,
		Type:        TypeTransfer,
		Amount:      req.Amount,
		Currency:    currency,
		Date:        date,
		Description: req.Description,
		Notes:       req.Notes,
	}
	dest := &Transaction{
		UserID:      userID,
		AccountID:   toID,
		Type:        TypeTransfer,
		Amount:      req.Amount,
		Currency:    currency,
		Date:        date,
		Description: req.Description,
		Notes:       req.Notes,
	}

	if err := s.repo.CreateTransfer(userID, source, dest); err != nil {
		return nil, err
	}
	return &TransferResult{Source: source, Dest: dest}, nil
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Now(), nil
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %q", s)
}

func currencyOrDefault(c, fallback string) string {
	if c == "" {
		return fallback
	}
	return c
}