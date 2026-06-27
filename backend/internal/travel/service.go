package travel

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserLookup is the contract for resolving user IDs (e.g. when adding a
// member by email). Implementations should return ErrUserNotFound when the
// user does not exist.
type UserLookup interface {
	FindByEmail(email string) (uuid.UUID, error)
}

type Service struct {
	repo Repository
	users UserLookup
}

func NewService(repo Repository, users UserLookup) *Service {
	return &Service{repo: repo, users: users}
}

var (
	ErrInvalidRole     = fmt.Errorf("invalid role")
	ErrUserNotFound    = fmt.Errorf("user not found")
	ErrSelfNotMember   = fmt.Errorf("cannot leave a group you own; transfer ownership or delete the group")
	ErrInvalidCurrency = fmt.Errorf("invalid currency code")
	ErrInvalidDate     = fmt.Errorf("invalid date")
)

// --- Groups ---

type CreateGroupRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Currency    string  `json:"currency"`
}

func (s *Service) CreateGroup(userID uuid.UUID, req CreateGroupRequest) (*TravelGroup, error) {
	if req.Currency == "" {
		req.Currency = "COP"
	}
	if len(req.Currency) != 3 {
		return nil, ErrInvalidCurrency
	}
	group := &TravelGroup{
		Name:        req.Name,
		Description: req.Description,
		Currency:    req.Currency,
		CreatedBy:   userID,
	}
	if err := s.repo.CreateGroup(group); err != nil {
		return nil, err
	}
	// Creator joins as owner.
	if err := s.repo.AddMember(&TravelGroupMember{
		GroupID: group.ID,
		UserID:  userID,
		Role:    RoleOwner,
	}); err != nil {
		return nil, err
	}
	return group, nil
}

func (s *Service) GetGroup(id, callerID uuid.UUID) (*TravelGroup, error) {
	if err := s.requireMembership(id, callerID); err != nil {
		return nil, err
	}
	return s.repo.GetGroup(id)
}

func (s *Service) ListGroups(userID uuid.UUID) ([]TravelGroup, error) {
	return s.repo.ListGroupsByUser(userID)
}

type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

func (s *Service) UpdateGroup(id, callerID uuid.UUID, req UpdateGroupRequest) (*TravelGroup, error) {
	if err := s.requireOwner(id, callerID); err != nil {
		return nil, err
	}
	g, err := s.repo.GetGroup(id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		g.Name = *req.Name
	}
	if req.Description != nil {
		g.Description = req.Description
	}
	if err := s.repo.UpdateGroup(g); err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) DeleteGroup(id, callerID uuid.UUID) error {
	if err := s.requireOwner(id, callerID); err != nil {
		return err
	}
	return s.repo.DeleteGroup(id)
}

// --- Members ---

type AddMemberRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (s *Service) AddMemberByEmail(groupID, callerID uuid.UUID, req AddMemberRequest) (*TravelGroupMember, error) {
	if err := s.requireOwner(groupID, callerID); err != nil {
		return nil, err
	}
	role := MemberRole(req.Role)
	if role == "" {
		role = RoleMember
	}
	if !IsValidRole(string(role)) {
		return nil, ErrInvalidRole
	}
	if s.users == nil {
		return nil, ErrUserNotFound
	}
	userID, err := s.users.FindByEmail(req.Email)
	if err != nil {
		return nil, ErrUserNotFound
	}
	// Idempotent: if already a member, return the existing record.
	if existing, err := s.repo.GetMember(groupID, userID); err == nil {
		return existing, nil
	}
	m := &TravelGroupMember{
		GroupID: groupID,
		UserID:  userID,
		Role:    role,
	}
	if err := s.repo.AddMember(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) ListMembers(groupID, callerID uuid.UUID) ([]TravelGroupMember, error) {
	if err := s.requireMembership(groupID, callerID); err != nil {
		return nil, err
	}
	return s.repo.ListMembers(groupID)
}

func (s *Service) RemoveMember(groupID, targetUserID, callerID uuid.UUID) error {
	member, err := s.repo.GetMember(groupID, targetUserID)
	if err != nil {
		return err
	}
	// Self-removal is always allowed. Removing someone else requires owner.
	if targetUserID != callerID {
		if err := s.requireOwner(groupID, callerID); err != nil {
			return err
		}
	}
	if member.Role == RoleOwner {
		owners, err := s.repo.CountOwners(groupID)
		if err != nil {
			return err
		}
		if owners <= 1 {
			return ErrCannotRemoveOwner
		}
	}
	return s.repo.RemoveMember(groupID, targetUserID)
}

// --- Expenses ---

type ShareInput struct {
	UserID string `json:"user_id"`
	// Amount is interpreted depending on SplitMethod:
	//   equal      — ignored, computed by service
	//   exact      — cents owed (raw value)
	//   percentage — basis points 0..10000 (so 25.5% is sent as 2550)
	Amount int64 `json:"amount"`
}

type AddExpenseRequest struct {
	PaidBy      string       `json:"paid_by"`
	Amount      int64        `json:"amount"`
	Currency    string       `json:"currency"`
	Description string       `json:"description"`
	SplitMethod string       `json:"split_method"`
	Date        string       `json:"date"`
	Shares      []ShareInput `json:"shares"`
}

func (s *Service) AddExpense(groupID, callerID uuid.UUID, req AddExpenseRequest) (*TravelExpense, []TravelExpenseShare, error) {
	if err := s.requireMembership(groupID, callerID); err != nil {
		return nil, nil, err
	}
	if req.Amount <= 0 {
		return nil, nil, fmt.Errorf("amount must be greater than zero")
	}
	if req.SplitMethod == "" {
		req.SplitMethod = string(SplitEqual)
	}
	if !IsValidSplitMethod(req.SplitMethod) {
		return nil, nil, ErrInvalidSplit
	}
	paidBy, err := uuid.Parse(req.PaidBy)
	if err != nil {
		return nil, nil, fmt.Errorf("parse paid_by: %w", err)
	}
	// Payer must be a member of the group.
	if _, err := s.repo.GetMember(groupID, paidBy); err != nil {
		return nil, nil, ErrPayerNotMember
	}
	if req.Currency == "" {
		g, err := s.repo.GetGroup(groupID)
		if err != nil {
			return nil, nil, err
		}
		req.Currency = g.Currency
	}

	shares, err := s.resolveShares(req.SplitMethod, req.Amount, req.Shares)
	if err != nil {
		return nil, nil, err
	}
	if len(shares) == 0 {
		return nil, nil, ErrSplitUsersMissing
	}
	// Every share user must be a member of the group.
	for _, sh := range shares {
		if _, err := s.repo.GetMember(groupID, sh.UserID); err != nil {
			return nil, nil, ErrNotMember
		}
	}

	date := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			date = t
		} else if t, err := time.Parse(time.RFC3339, req.Date); err == nil {
			date = t
		} else {
			return nil, nil, ErrInvalidDate
		}
	}

	exp := &TravelExpense{
		GroupID:     groupID,
		PaidBy:      paidBy,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		SplitMethod: SplitMethod(req.SplitMethod),
		Date:        date,
	}
	if err := s.repo.CreateExpenseWithShares(exp, shares); err != nil {
		return nil, nil, err
	}
	return exp, shares, nil
}

func (s *Service) ListExpenses(groupID, callerID uuid.UUID) ([]TravelExpense, error) {
	if err := s.requireMembership(groupID, callerID); err != nil {
		return nil, err
	}
	return s.repo.ListExpensesByGroup(groupID)
}

func (s *Service) GetExpense(expenseID, callerID uuid.UUID) (*TravelExpense, []TravelExpenseShare, error) {
	exp, err := s.repo.GetExpense(expenseID)
	if err != nil {
		return nil, nil, err
	}
	if err := s.requireMembership(exp.GroupID, callerID); err != nil {
		return nil, nil, err
	}
	shares, err := s.repo.ListSharesByExpense(expenseID)
	if err != nil {
		return nil, nil, err
	}
	return exp, shares, nil
}

func (s *Service) DeleteExpense(expenseID, callerID uuid.UUID) error {
	exp, err := s.repo.GetExpense(expenseID)
	if err != nil {
		return err
	}
	if err := s.requireMembership(exp.GroupID, callerID); err != nil {
		return err
	}
	return s.repo.DeleteExpense(expenseID)
}

// --- Settlements ---

// SettlementSuggestion is one row in the "who owes whom" output.
type SettlementSuggestion struct {
	FromUser uuid.UUID `json:"from_user"`
	ToUser   uuid.UUID `json:"to_user"`
	Amount   int64     `json:"amount"`
}

// ComputeSettlements runs the greedy minimum-transfer algorithm to produce a
// list of (from -> to, amount) tuples that, when executed, will zero out
// every member's balance in the group.
//
// Algorithm:
//  1. For each member, balance = paid - owed.
//  2. Sort members by balance descending; build a min-heap of debtors.
//  3. Repeatedly take the largest creditor and the largest debtor, transfer
//     min(|debtor|, creditor), and re-insert any leftover.
func (s *Service) ComputeSettlements(groupID, callerID uuid.UUID) ([]SettlementSuggestion, error) {
	if err := s.requireMembership(groupID, callerID); err != nil {
		return nil, err
	}
	members, err := s.repo.ListMembers(groupID)
	if err != nil {
		return nil, err
	}
	balances := make(map[uuid.UUID]int64, len(members))
	for _, m := range members {
		balances[m.UserID] = 0
	}
	for userID := range balances {
		paid, err := s.repo.SumPaidByUser(groupID, userID)
		if err != nil {
			return nil, err
		}
		owed, err := s.repo.SumShareByUser(groupID, userID)
		if err != nil {
			return nil, err
		}
		balances[userID] = paid - owed
	}

	creditors := make([]balanceEntry, 0)
	debtors := make([]balanceEntry, 0)
	for userID, bal := range balances {
		if bal > 0 {
			creditors = append(creditors, balanceEntry{userID: userID, balance: bal})
		} else if bal < 0 {
			debtors = append(debtors, balanceEntry{userID: userID, balance: -bal})
		}
	}
	// Sort: largest first on each side.
	sortEntriesDesc(creditors)
	sortEntriesDesc(debtors)

	suggestions := make([]SettlementSuggestion, 0)
	ci, di := 0, 0
	for ci < len(creditors) && di < len(debtors) {
		c := creditors[ci]
		d := debtors[di]
		transfer := c.balance
		if d.balance < transfer {
			transfer = d.balance
		}
		if transfer > 0 {
			suggestions = append(suggestions, SettlementSuggestion{
				FromUser: d.userID,
				ToUser:   c.userID,
				Amount:   transfer,
			})
		}
		creditors[ci].balance -= transfer
		debtors[di].balance -= transfer
		if creditors[ci].balance == 0 {
			ci++
		}
		if debtors[di].balance == 0 {
			di++
		}
	}
	return suggestions, nil
}

type balanceEntry struct {
	userID  uuid.UUID
	balance int64
}

func sortEntriesDesc(s []balanceEntry) {
	// Simple insertion sort — group sizes are small (typically < 20).
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].balance > s[j-1].balance; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

type RecordSettlementRequest struct {
	FromUser string `json:"from_user"`
	ToUser   string `json:"to_user"`
	Amount   int64  `json:"amount"`
}

// RecordSettlement creates a pending settlement record (a snapshot of the
// debt that the group agreed to settle). It does NOT modify balances — those
// remain tied to expenses until the group resets them.
func (s *Service) RecordSettlement(groupID, callerID uuid.UUID, req RecordSettlementRequest) (*TravelSettlement, error) {
	if err := s.requireMembership(groupID, callerID); err != nil {
		return nil, err
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than zero")
	}
	from, err := uuid.Parse(req.FromUser)
	if err != nil {
		return nil, fmt.Errorf("parse from_user: %w", err)
	}
	to, err := uuid.Parse(req.ToUser)
	if err != nil {
		return nil, fmt.Errorf("parse to_user: %w", err)
	}
	if from == to {
		return nil, fmt.Errorf("from_user and to_user must differ")
	}
	if _, err := s.repo.GetMember(groupID, from); err != nil {
		return nil, ErrNotMember
	}
	if _, err := s.repo.GetMember(groupID, to); err != nil {
		return nil, ErrNotMember
	}
	g, err := s.repo.GetGroup(groupID)
	if err != nil {
		return nil, err
	}
	rec := &TravelSettlement{
		GroupID:  groupID,
		FromUser: from,
		ToUser:   to,
		Amount:   req.Amount,
		Currency: g.Currency,
		Status:   SettlementPending,
	}
	if err := s.repo.CreateSettlement(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *Service) ListSettlements(groupID, callerID uuid.UUID) ([]TravelSettlement, error) {
	if err := s.requireMembership(groupID, callerID); err != nil {
		return nil, err
	}
	return s.repo.ListSettlementsByGroup(groupID)
}

func (s *Service) ConfirmSettlement(settlementID, callerID uuid.UUID) (*TravelSettlement, error) {
	rec, err := s.repo.GetSettlement(settlementID)
	if err != nil {
		return nil, err
	}
	if err := s.requireMembership(rec.GroupID, callerID); err != nil {
		return nil, err
	}
	// Only the recipient (to_user) can confirm a payment was made.
	if rec.ToUser != callerID {
		return nil, fmt.Errorf("only the recipient can confirm a settlement")
	}
	if rec.Status == SettlementConfirmed {
		return rec, nil
	}
	now := time.Now()
	rec.Status = SettlementConfirmed
	rec.ConfirmedAt = &now
	if err := s.repo.UpdateSettlement(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

// --- helpers ---

func (s *Service) requireMembership(groupID, userID uuid.UUID) error {
	if _, err := s.repo.GetMember(groupID, userID); err != nil {
		return ErrNotMember
	}
	return nil
}

func (s *Service) requireOwner(groupID, userID uuid.UUID) error {
	m, err := s.repo.GetMember(groupID, userID)
	if err != nil {
		return ErrNotMember
	}
	if m.Role != RoleOwner {
		return fmt.Errorf("owner role required")
	}
	return nil
}

// resolveShares computes the final share amounts based on the chosen method.
// For SplitEqual the caller's Shares slice is replaced with one row per user,
// each carrying amount = total / len. The leftover cents (from integer
// division) are distributed one cent at a time to the first N members so the
// sum always equals the total exactly.
//
// For SplitExact, the caller's amounts are taken verbatim and must sum to
// the expense total.
//
// For SplitPercentage, the caller's amounts are interpreted as basis points
// (so 25% is sent as 2500). They must sum to exactly 10000.
func (s *Service) resolveShares(method string, total int64, inputs []ShareInput) ([]TravelExpenseShare, error) {
	switch SplitMethod(method) {
	case SplitEqual:
		if len(inputs) == 0 {
			return nil, ErrSplitUsersMissing
		}
		base := total / int64(len(inputs))
		leftover := total - base*int64(len(inputs))
		out := make([]TravelExpenseShare, 0, len(inputs))
		for i, in := range inputs {
			uid, err := uuid.Parse(in.UserID)
			if err != nil {
				return nil, fmt.Errorf("parse user_id: %w", err)
			}
			amt := base
			if int64(i) < leftover {
				amt++
			}
			out = append(out, TravelExpenseShare{UserID: uid, Amount: amt})
		}
		return out, nil

	case SplitExact:
		if len(inputs) == 0 {
			return nil, ErrSplitUsersMissing
		}
		var sum int64
		out := make([]TravelExpenseShare, 0, len(inputs))
		for _, in := range inputs {
			uid, err := uuid.Parse(in.UserID)
			if err != nil {
				return nil, fmt.Errorf("parse user_id: %w", err)
			}
			out = append(out, TravelExpenseShare{UserID: uid, Amount: in.Amount})
			sum += in.Amount
		}
		if sum != total {
			return nil, fmt.Errorf("%w: exact split sum=%d, expected=%d", ErrSplitSumMismatch, sum, total)
		}
		return out, nil

	case SplitPercentage:
		if len(inputs) == 0 {
			return nil, ErrSplitUsersMissing
		}
		var bpsTotal int64
		out := make([]TravelExpenseShare, 0, len(inputs))
		for _, in := range inputs {
			uid, err := uuid.Parse(in.UserID)
			if err != nil {
				return nil, fmt.Errorf("parse user_id: %w", err)
			}
			out = append(out, TravelExpenseShare{UserID: uid, Amount: 0}) // placeholder
			bpsTotal += in.Amount
		}
		if bpsTotal != 10000 {
			return nil, fmt.Errorf("%w: percentage basis points sum=%d, expected=10000", ErrSplitSumMismatch, bpsTotal)
		}
		// Distribute amounts. We compute raw = total * bps / 10000 and assign
		// the leftover cents to the first N members (mirrors the equal-split
		// rounding strategy).
		var assigned int64
		for i := range out {
			bps := inputs[i].Amount
			raw := total * bps / 10000
			out[i].Amount = raw
			assigned += raw
		}
		leftover := total - assigned
		idx := 0
		for leftover > 0 && idx < len(out) {
			out[idx].Amount++
			leftover--
			idx++
		}
		return out, nil
	}
	return nil, ErrInvalidSplit
}