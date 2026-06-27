package accounts

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidType = errors.New("invalid account type")

type Service interface {
	Create(userID uuid.UUID, req CreateRequest) (*Account, error)
	List(userID uuid.UUID) ([]Account, error)
	Get(id, userID uuid.UUID) (*Account, error)
	Update(id, userID uuid.UUID, req UpdateRequest) (*Account, error)
	Delete(id, userID uuid.UUID) error
}

type service struct {
	repo AccountRepository
}

func NewService(repo AccountRepository) Service {
	return &service{repo: repo}
}

func (s *service) Create(userID uuid.UUID, req CreateRequest) (*Account, error) {
	if !IsValidType(req.Type) {
		return nil, ErrInvalidType
	}
	a := &Account{
		UserID:         userID,
		Name:           req.Name,
		Type:           AccountType(req.Type),
		Currency:       req.Currency,
		OpeningBalance: req.OpeningBalance,
		Color:          req.Color,
		Icon:           req.Icon,
	}
	if err := s.repo.Create(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) List(userID uuid.UUID) ([]Account, error) {
	return s.repo.ListByUser(userID)
}

func (s *service) Get(id, userID uuid.UUID) (*Account, error) {
	return s.repo.GetByID(id, userID)
}

func (s *service) Update(id, userID uuid.UUID, req UpdateRequest) (*Account, error) {
	a, err := s.repo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		a.Name = *req.Name
	}
	if req.Color != nil {
		a.Color = req.Color
	}
	if req.Icon != nil {
		a.Icon = req.Icon
	}
	if err := s.repo.Update(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *service) Delete(id, userID uuid.UUID) error {
	return s.repo.Delete(id, userID)
}

func IsValidType(t string) bool {
	switch AccountType(t) {
	case TypeCash, TypeDebit, TypeCredit, TypeSavings:
		return true
	}
	return false
}