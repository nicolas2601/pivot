package goals

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAmount   = errors.New("target_amount must be greater than zero")
	ErrInvalidName     = errors.New("name is required")
	ErrInvalidCurrency = errors.New("currency must be a 3-letter ISO 4217 code")
	ErrInvalidDeadline = errors.New("deadline must be in the future")
	ErrNonPositiveMove = errors.New("deposit/withdraw amount must be greater than zero")
	ErrOverWithdraw    = errors.New("withdraw would push current_amount below zero")
	ErrGoalAlreadyDone = errors.New("goal already completed; deposit or re-open first")
)

// AccountLookup validates that the optional account_id exists & belongs to user.
// Returns (false, nil) when accountID is nil.
type AccountLookup interface {
	Exists(id, userID uuid.UUID) (bool, error)
}

type Service struct {
	repo     Repository
	accounts AccountLookup
}

func NewService(repo Repository, accounts AccountLookup) *Service {
	return &Service{repo: repo, accounts: accounts}
}

type CreateRequest struct {
	Name          string  `json:"name"`
	TargetAmount  int64   `json:"target_amount"`
	Currency      string  `json:"currency"`
	Deadline      *string `json:"deadline,omitempty"`
	AccountID     *string `json:"account_id,omitempty"`
	Color         *string `json:"color,omitempty"`
	Notes         *string `json:"notes,omitempty"`
}

type UpdateRequest struct {
	Name         *string `json:"name,omitempty"`
	TargetAmount *int64  `json:"target_amount,omitempty"`
	Deadline     *string `json:"deadline,omitempty"`
	ClearDeadline bool   `json:"clear_deadline,omitempty"`
	AccountID    *string `json:"account_id,omitempty"`
	ClearAccount  bool   `json:"clear_account,omitempty"`
	Color        *string `json:"color,omitempty"`
	Notes        *string `json:"notes,omitempty"`
}

type MoveRequest struct {
	Amount int64  `json:"amount"`
	Note   string `json:"note,omitempty"`
}

func (s *Service) Create(userID uuid.UUID, req CreateRequest) (*Goal, error) {
	if req.Name == "" {
		return nil, ErrInvalidName
	}
	if req.TargetAmount <= 0 {
		return nil, ErrInvalidAmount
	}
	currency := req.Currency
	if currency == "" {
		currency = "COP"
	}
	if len(currency) != 3 {
		return nil, ErrInvalidCurrency
	}
	deadline, err := parseOptionalDate(req.Deadline)
	if err != nil {
		return nil, ErrInvalidDeadline
	}
	if deadline != nil && deadline.Before(time.Now().Truncate(24 * time.Hour)) {
		return nil, ErrInvalidDeadline
	}
	var accountID *uuid.UUID
	if req.AccountID != nil && *req.AccountID != "" {
		aid, err := uuid.Parse(*req.AccountID)
		if err != nil {
			return nil, fmt.Errorf("parse account_id: %w", err)
		}
		if s.accounts != nil {
			ok, err := s.accounts.Exists(aid, userID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("account not found or not owned by user")
			}
		}
		accountID = &aid
	}

	g := &Goal{
		UserID:       userID,
		Name:         req.Name,
		TargetAmount: req.TargetAmount,
		Currency:     currency,
		Deadline:     deadline,
		AccountID:    accountID,
		Color:        req.Color,
		Notes:        req.Notes,
	}
	if err := s.repo.Create(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) Get(id, userID uuid.UUID) (*Goal, error) {
	return s.repo.GetByID(id, userID)
}

func (s *Service) List(userID uuid.UUID) ([]Goal, error) {
	return s.repo.ListByUser(userID)
}

func (s *Service) Update(id, userID uuid.UUID, req UpdateRequest) (*Goal, error) {
	g, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		if *req.Name == "" {
			return nil, ErrInvalidName
		}
		g.Name = *req.Name
	}
	if req.TargetAmount != nil {
		if *req.TargetAmount <= 0 {
			return nil, ErrInvalidAmount
		}
		g.TargetAmount = *req.TargetAmount
	}
	if req.ClearDeadline {
		g.Deadline = nil
	} else if req.Deadline != nil && *req.Deadline != "" {
		d, err := parseOptionalDate(req.Deadline)
		if err != nil {
			return nil, ErrInvalidDeadline
		}
		g.Deadline = d
	}
	if req.ClearAccount {
		g.AccountID = nil
	} else if req.AccountID != nil && *req.AccountID != "" {
		aid, err := uuid.Parse(*req.AccountID)
		if err != nil {
			return nil, fmt.Errorf("parse account_id: %w", err)
		}
		if s.accounts != nil {
			ok, err := s.accounts.Exists(aid, userID)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("account not found or not owned by user")
			}
		}
		g.AccountID = &aid
	}
	if req.Color != nil {
		g.Color = req.Color
	}
	if req.Notes != nil {
		g.Notes = req.Notes
	}
	if err := s.repo.Update(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

// Deposit adds funds to a goal. Atomic — safe under concurrent calls.
func (s *Service) Deposit(id, userID uuid.UUID, req MoveRequest) (*Goal, error) {
	if req.Amount <= 0 {
		return nil, ErrNonPositiveMove
	}
	return s.repo.AtomicAdjustCurrent(id, userID, req.Amount)
}

// Withdraw removes funds from a goal. Rejected if it would go negative.
func (s *Service) Withdraw(id, userID uuid.UUID, req MoveRequest) (*Goal, error) {
	if req.Amount <= 0 {
		return nil, ErrNonPositiveMove
	}
	current, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	if current.CurrentAmount-req.Amount < 0 {
		return nil, ErrOverWithdraw
	}
	return s.repo.AtomicAdjustCurrent(id, userID, -req.Amount)
}

func parseOptionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	if t, err := time.Parse("2006-01-02", *s); err == nil {
		return &t, nil
	}
	if t, err := time.Parse(time.RFC3339, *s); err == nil {
		return &t, nil
	}
	return nil, fmt.Errorf("invalid date format: %s", *s)
}
