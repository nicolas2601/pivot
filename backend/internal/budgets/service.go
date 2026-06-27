package budgets

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CategoryLookup is the contract for validating that a category exists and is
// owned by the user. Implementations should return ErrCategoryNotFound (or
// nil/false if the contract returns a bool) when the category is unknown.
type CategoryLookup interface {
	GetByID(id, userID uuid.UUID) (exists bool, err error)
}

type Service struct {
	repo      Repository
	categories CategoryLookup
}

func NewService(repo Repository, categories CategoryLookup) *Service {
	return &Service{repo: repo, categories: categories}
}

var (
	ErrInvalidPeriod   = fmt.Errorf("invalid period (must be monthly or yearly)")
	ErrInvalidAmount   = fmt.Errorf("amount must be greater than zero")
	ErrInvalidDate     = fmt.Errorf("start_date is required")
	ErrEndBeforeStart = fmt.Errorf("end_date must be on or after start_date")
	ErrCategoryMissing = fmt.Errorf("category not found or not owned by user")
)

type CreateRequest struct {
	CategoryID string  `json:"category_id"`
	Amount     int64   `json:"amount"`
	Period     string  `json:"period"`
	StartDate  string  `json:"start_date"`
	EndDate    *string `json:"end_date,omitempty"`
}

type UpdateRequest struct {
	Amount    *int64  `json:"amount,omitempty"`
	Period    *string `json:"period,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	EndDate   *string `json:"end_date,omitempty"`
	ClearEnd  bool    `json:"clear_end_date,omitempty"`
}

func (s *Service) Create(userID uuid.UUID, req CreateRequest) (*Budget, error) {
	if !IsValidPeriod(req.Period) {
		return nil, ErrInvalidPeriod
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	start, err := parseDate(req.StartDate)
	if err != nil {
		return nil, ErrInvalidDate
	}
	var end *time.Time
	if req.EndDate != nil && *req.EndDate != "" {
		e, err := parseDate(*req.EndDate)
		if err != nil {
			return nil, ErrInvalidDate
		}
		if e.Before(start) {
			return nil, ErrEndBeforeStart
		}
		end = &e
	}

	cid, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("parse category_id: %w", err)
	}
	if s.categories != nil {
		exists, err := s.categories.GetByID(cid, userID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrCategoryMissing
		}
	}

	b := &Budget{
		UserID:     userID,
		CategoryID: cid,
		Amount:     req.Amount,
		Period:     Period(req.Period),
		StartDate:  start,
		EndDate:    end,
	}
	if err := s.repo.Create(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) Get(id, userID uuid.UUID) (*Budget, error) {
	return s.repo.GetByID(id, userID)
}

func (s *Service) List(userID uuid.UUID) ([]Budget, error) {
	return s.repo.ListByUser(userID)
}

func (s *Service) Update(id, userID uuid.UUID, req UpdateRequest) (*Budget, error) {
	b, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	if req.Amount != nil {
		if *req.Amount <= 0 {
			return nil, ErrInvalidAmount
		}
		b.Amount = *req.Amount
	}
	if req.Period != nil {
		if !IsValidPeriod(*req.Period) {
			return nil, ErrInvalidPeriod
		}
		b.Period = Period(*req.Period)
	}
	if req.StartDate != nil {
		start, err := parseDate(*req.StartDate)
		if err != nil {
			return nil, ErrInvalidDate
		}
		b.StartDate = start
	}
	if req.ClearEnd {
		b.EndDate = nil
	} else if req.EndDate != nil && *req.EndDate != "" {
		e, err := parseDate(*req.EndDate)
		if err != nil {
			return nil, ErrInvalidDate
		}
		if e.Before(b.StartDate) {
			return nil, ErrEndBeforeStart
		}
		b.EndDate = &e
	}
	if err := s.repo.Update(b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, ErrInvalidDate
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, ErrInvalidDate
}