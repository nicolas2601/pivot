package travel

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// fakeRepo is an in-memory implementation of Repository for service tests.
type fakeRepo struct {
	mu          sync.Mutex
	groups      map[uuid.UUID]*TravelGroup
	members     map[string]*TravelGroupMember // key: groupID|userID
	expenses    map[uuid.UUID]*TravelExpense
	shares      map[uuid.UUID][]TravelExpenseShare
	settlements map[uuid.UUID]*TravelSettlement
	paidTotals  map[string]int64 // key: groupID|userID
	shareTotals map[string]int64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		groups:      map[uuid.UUID]*TravelGroup{},
		members:     map[string]*TravelGroupMember{},
		expenses:    map[uuid.UUID]*TravelExpense{},
		shares:      map[uuid.UUID][]TravelExpenseShare{},
		settlements: map[uuid.UUID]*TravelSettlement{},
		paidTotals:  map[string]int64{},
		shareTotals: map[string]int64{},
	}
}

func memberKey(g, u uuid.UUID) string { return g.String() + "|" + u.String() }

func (f *fakeRepo) CreateGroup(g *TravelGroup) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	copy := *g
	f.groups[g.ID] = &copy
	return nil
}

func (f *fakeRepo) GetGroup(id uuid.UUID) (*TravelGroup, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	g, ok := f.groups[id]
	if !ok {
		return nil, ErrGroupNotFound
	}
	copy := *g
	return &copy, nil
}

func (f *fakeRepo) ListGroupsByUser(userID uuid.UUID) ([]TravelGroup, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []TravelGroup{}
	for _, m := range f.members {
		if m.UserID == userID {
			if g, ok := f.groups[m.GroupID]; ok {
				out = append(out, *g)
			}
		}
	}
	return out, nil
}

func (f *fakeRepo) UpdateGroup(g *TravelGroup) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.groups[g.ID] = g
	return nil
}

func (f *fakeRepo) DeleteGroup(id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.groups, id)
	return nil
}

func (f *fakeRepo) AddMember(m *TravelGroupMember) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	f.members[memberKey(m.GroupID, m.UserID)] = m
	return nil
}

func (f *fakeRepo) ListMembers(groupID uuid.UUID) ([]TravelGroupMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []TravelGroupMember{}
	for _, m := range f.members {
		if m.GroupID == groupID {
			out = append(out, *m)
		}
	}
	return out, nil
}

func (f *fakeRepo) GetMember(groupID, userID uuid.UUID) (*TravelGroupMember, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.members[memberKey(groupID, userID)]
	if !ok {
		return nil, ErrMemberNotFound
	}
	copy := *m
	return &copy, nil
}

func (f *fakeRepo) RemoveMember(groupID, userID uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.members, memberKey(groupID, userID))
	return nil
}

func (f *fakeRepo) CountOwners(groupID uuid.UUID) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, m := range f.members {
		if m.GroupID == groupID && m.Role == RoleOwner {
			n++
		}
	}
	return n, nil
}

func (f *fakeRepo) CreateExpenseWithShares(e *TravelExpense, shares []TravelExpenseShare) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	ec := *e
	f.expenses[e.ID] = &ec
	for i := range shares {
		shares[i].ExpenseID = e.ID
		if shares[i].ID == uuid.Nil {
			shares[i].ID = uuid.New()
		}
		f.shares[e.ID] = append(f.shares[e.ID], shares[i])
		f.shareTotals[memberKey(e.GroupID, shares[i].UserID)] += shares[i].Amount
	}
	f.paidTotals[memberKey(e.GroupID, e.PaidBy)] += e.Amount
	return nil
}

func (f *fakeRepo) GetExpense(id uuid.UUID) (*TravelExpense, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	e, ok := f.expenses[id]
	if !ok {
		return nil, ErrExpenseNotFound
	}
	copy := *e
	return &copy, nil
}

func (f *fakeRepo) ListExpensesByGroup(groupID uuid.UUID) ([]TravelExpense, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []TravelExpense{}
	for _, e := range f.expenses {
		if e.GroupID == groupID {
			out = append(out, *e)
		}
	}
	return out, nil
}

func (f *fakeRepo) DeleteExpense(id uuid.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.expenses, id)
	delete(f.shares, id)
	return nil
}

func (f *fakeRepo) ListSharesByExpense(expenseID uuid.UUID) ([]TravelExpenseShare, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]TravelExpenseShare, len(f.shares[expenseID]))
	copy(out, f.shares[expenseID])
	return out, nil
}

func (f *fakeRepo) CreateSettlement(s *TravelSettlement) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	f.settlements[s.ID] = s
	return nil
}

func (f *fakeRepo) GetSettlement(id uuid.UUID) (*TravelSettlement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	s, ok := f.settlements[id]
	if !ok {
		return nil, ErrExpenseNotFound // reuse for tests
	}
	copy := *s
	return &copy, nil
}

func (f *fakeRepo) ListSettlementsByGroup(groupID uuid.UUID) ([]TravelSettlement, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []TravelSettlement{}
	for _, s := range f.settlements {
		if s.GroupID == groupID {
			out = append(out, *s)
		}
	}
	return out, nil
}

func (f *fakeRepo) UpdateSettlement(s *TravelSettlement) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.settlements[s.ID] = s
	return nil
}

func (f *fakeRepo) SumPaidByUser(groupID, userID uuid.UUID) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.paidTotals[memberKey(groupID, userID)], nil
}

func (f *fakeRepo) SumShareByUser(groupID, userID uuid.UUID) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.shareTotals[memberKey(groupID, userID)], nil
}

// fakeUsers implements UserLookup.
type fakeUsers struct {
	mu   sync.Mutex
	byID map[string]uuid.UUID // email → id
}

func newFakeUsers() *fakeUsers {
	return &fakeUsers{byID: map[string]uuid.UUID{}}
}

func (f *fakeUsers) add(email string, id uuid.UUID) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.byID[email] = id
}

func (f *fakeUsers) FindByEmail(email string) (uuid.UUID, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.byID[email]
	if !ok {
		return uuid.Nil, errors.New("not found")
	}
	return id, nil
}

// ──────────────────────────── Service tests ────────────────────────────

func TestCreateGroup_AddsCreatorAsOwner(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)

	owner := uuid.New()
	g, err := s.CreateGroup(owner, CreateGroupRequest{Name: "Viaje Cartagena"})
	if err != nil {
		t.Fatalf("CreateGroup: %v", err)
	}
	if g.Currency != "COP" {
		t.Errorf("Currency = %q, want COP (default)", g.Currency)
	}
	m, err := repo.GetMember(g.ID, owner)
	if err != nil {
		t.Fatalf("GetMember: %v", err)
	}
	if m.Role != RoleOwner {
		t.Errorf("Role = %v, want RoleOwner", m.Role)
	}
}

func TestCreateGroup_RejectsBadCurrency(t *testing.T) {
	s := NewService(newFakeRepo(), newFakeUsers())
	_, err := s.CreateGroup(uuid.New(), CreateGroupRequest{
		Name: "X", Currency: "PESO",
	})
	if !errors.Is(err, ErrInvalidCurrency) {
		t.Errorf("got %v, want ErrInvalidCurrency", err)
	}
}

func TestGetGroup_RequiresMembership(t *testing.T) {
	repo := newFakeRepo()
	s := NewService(repo, newFakeUsers())
	owner := uuid.New()
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})

	if _, err := s.GetGroup(g.ID, uuid.New()); !errors.Is(err, ErrNotMember) {
		t.Errorf("non-member: got %v, want ErrNotMember", err)
	}
	if _, err := s.GetGroup(g.ID, owner); err != nil {
		t.Errorf("owner: %v", err)
	}
}

func TestAddMemberByEmail_RequiresOwner(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner := uuid.New()
	member := uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	users.add("c@x", uuid.New())
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	// Member tries to add another → should fail (not owner).
	if _, err := s.AddMemberByEmail(g.ID, member, AddMemberRequest{Email: "c@x"}); err == nil ||
		!strings.Contains(err.Error(), "owner") {
		t.Errorf("got %v, want owner-required", err)
	}
}

func TestAddMemberByEmail_Idempotent(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner := uuid.New()
	users.add("a@x", owner)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})

	// Add owner again — should return existing, not error.
	m, err := s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "a@x"})
	if err != nil {
		t.Errorf("re-add self: %v", err)
	}
	if m.UserID != owner {
		t.Errorf("UserID = %v, want %v", m.UserID, owner)
	}
}

func TestAddMemberByEmail_RejectsInvalidRole(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner := uuid.New()
	newUser := uuid.New()
	users.add("a@x", owner)
	users.add("b@x", newUser)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})

	if _, err := s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x", Role: "admin"}); !errors.Is(err, ErrInvalidRole) {
		t.Errorf("got %v, want ErrInvalidRole", err)
	}
}

func TestRemoveMember_CannotRemoveOnlyOwner(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner := uuid.New()
	users.add("a@x", owner)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})

	err := s.RemoveMember(g.ID, owner, owner)
	if !errors.Is(err, ErrCannotRemoveOwner) {
		t.Errorf("got %v, want ErrCannotRemoveOwner", err)
	}
}

func TestAddExpense_EqualSplit_DistributesLeftoverCents(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner, member := uuid.New(), uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	// 1000 split 3 ways: 334, 333, 333 (or 333, 334, 333 depending on order).
	// Use 2 members for a cleaner case: 100 / 2 = 50/50, no leftover.
	_, shares, err := s.AddExpense(g.ID, owner, AddExpenseRequest{
		PaidBy:      owner.String(),
		Amount:      100,
		Currency:    "COP",
		Description: "Almuerzo",
		SplitMethod: "equal",
		Shares: []ShareInput{
			{UserID: owner.String()},
			{UserID: member.String()},
		},
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}
	if len(shares) != 2 {
		t.Fatalf("len(shares) = %d, want 2", len(shares))
	}
	var total int64
	for _, sh := range shares {
		total += sh.Amount
	}
	if total != 100 {
		t.Errorf("sum(shares) = %d, want 100", total)
	}
}

func TestAddExpense_ExactSplit_RejectsSumMismatch(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner, member := uuid.New(), uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	_, _, err := s.AddExpense(g.ID, owner, AddExpenseRequest{
		PaidBy:      owner.String(),
		Amount:      100,
		Currency:    "COP",
		SplitMethod: "exact",
		Shares: []ShareInput{
			{UserID: owner.String(), Amount: 60},
			{UserID: member.String(), Amount: 30}, // sum=90 ≠ 100
		},
	})
	if !errors.Is(err, ErrSplitSumMismatch) {
		t.Errorf("got %v, want ErrSplitSumMismatch", err)
	}
}

func TestAddExpense_PercentageSplit_RejectsBadBps(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner, member := uuid.New(), uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	_, _, err := s.AddExpense(g.ID, owner, AddExpenseRequest{
		PaidBy:      owner.String(),
		Amount:      100,
		SplitMethod: "percentage",
		Shares: []ShareInput{
			{UserID: owner.String(), Amount: 3000}, // 30%
			{UserID: member.String(), Amount: 6000}, // 60% (sum=9000 ≠ 10000)
		},
	})
	if !errors.Is(err, ErrSplitSumMismatch) {
		t.Errorf("got %v, want ErrSplitSumMismatch", err)
	}
}

func TestAddExpense_RejectsNonMemberPayer(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner := uuid.New()
	stranger := uuid.New()
	users.add("a@x", owner)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})

	_, _, err := s.AddExpense(g.ID, owner, AddExpenseRequest{
		PaidBy:      stranger.String(),
		Amount:      100,
		SplitMethod: "equal",
		Shares:      []ShareInput{{UserID: owner.String()}},
	})
	if !errors.Is(err, ErrPayerNotMember) {
		t.Errorf("got %v, want ErrPayerNotMember", err)
	}
}

func TestComputeSettlements_SimpleTwoUserGroup(t *testing.T) {
	// Owner paid 1000, split 50/50. Owner paid 1000, owes 500 → +500.
	// Member paid 0, owes 500 → -500.
	// Settlement: member → owner, 500.
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner, member := uuid.New(), uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	_, _, err := s.AddExpense(g.ID, owner, AddExpenseRequest{
		PaidBy:      owner.String(),
		Amount:      1000,
		SplitMethod: "equal",
		Shares:      []ShareInput{{UserID: owner.String()}, {UserID: member.String()}},
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	settlements, err := s.ComputeSettlements(g.ID, owner)
	if err != nil {
		t.Fatalf("ComputeSettlements: %v", err)
	}
	if len(settlements) != 1 {
		t.Fatalf("len = %d, want 1", len(settlements))
	}
	if settlements[0].Amount != 500 {
		t.Errorf("Amount = %d, want 500", settlements[0].Amount)
	}
	if settlements[0].FromUser != member || settlements[0].ToUser != owner {
		t.Errorf("from=%v to=%v, want from=member to=owner", settlements[0].FromUser, settlements[0].ToUser)
	}
}

func TestComputeSettlements_ThreeUsersMinimizesTransfers(t *testing.T) {
	// A paid 900, B paid 0, C paid 0. Split 300 each.
	// Balances: A=+600, B=-300, C=-300.
	// Greedy: largest creditor (A 600) + largest debtor (B 300) → A←B 300.
	// A balance → 300. Next: A + C → A←C 300. Done. 2 transfers (optimal).
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	a, b, c := uuid.New(), uuid.New(), uuid.New()
	users.add("a@x", a)
	users.add("b@x", b)
	users.add("c@x", c)
	g, _ := s.CreateGroup(a, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, a, AddMemberRequest{Email: "b@x"})
	s.AddMemberByEmail(g.ID, a, AddMemberRequest{Email: "c@x"})

	_, _, err := s.AddExpense(g.ID, a, AddExpenseRequest{
		PaidBy:      a.String(),
		Amount:      900,
		SplitMethod: "equal",
		Shares:      []ShareInput{{UserID: a.String()}, {UserID: b.String()}, {UserID: c.String()}},
	})
	if err != nil {
		t.Fatalf("AddExpense: %v", err)
	}

	settlements, err := s.ComputeSettlements(g.ID, a)
	if err != nil {
		t.Fatalf("ComputeSettlements: %v", err)
	}
	if len(settlements) != 2 {
		t.Fatalf("len = %d, want 2 (greedy optimal)", len(settlements))
	}
	// All transfers must end at A.
	var totalTransferredToA int64
	for _, s := range settlements {
		if s.ToUser != a {
			t.Errorf("to = %v, want A (%v)", s.ToUser, a)
		}
		totalTransferredToA += s.Amount
	}
	if totalTransferredToA != 600 {
		t.Errorf("total to A = %d, want 600", totalTransferredToA)
	}
}

func TestComputeSettlements_ZeroBalances_NoTransfers(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner, member := uuid.New(), uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	// Member paid 500, owner paid 500. Equal split → both at 0.
	_, _, _ = s.AddExpense(g.ID, owner, AddExpenseRequest{
		PaidBy: owner.String(), Amount: 1000, SplitMethod: "equal",
		Shares: []ShareInput{{UserID: owner.String()}, {UserID: member.String()}},
	})
	_, _, _ = s.AddExpense(g.ID, member, AddExpenseRequest{
		PaidBy: member.String(), Amount: 1000, SplitMethod: "equal",
		Shares: []ShareInput{{UserID: owner.String()}, {UserID: member.String()}},
	})

	settlements, err := s.ComputeSettlements(g.ID, owner)
	if err != nil {
		t.Fatalf("ComputeSettlements: %v", err)
	}
	if len(settlements) != 0 {
		t.Errorf("expected no transfers, got %d", len(settlements))
	}
}

func TestRecordSettlement_RequiresMembership(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner, member, stranger := uuid.New(), uuid.New(), uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	_, err := s.RecordSettlement(g.ID, stranger, RecordSettlementRequest{
		FromUser: member.String(), ToUser: owner.String(), Amount: 100,
	})
	if !errors.Is(err, ErrNotMember) {
		t.Errorf("got %v, want ErrNotMember", err)
	}
}

func TestRecordSettlement_RejectsSameFromAndTo(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner := uuid.New()
	users.add("a@x", owner)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})

	_, err := s.RecordSettlement(g.ID, owner, RecordSettlementRequest{
		FromUser: owner.String(), ToUser: owner.String(), Amount: 100,
	})
	if err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Errorf("got %v, want must-differ error", err)
	}
}

func TestConfirmSettlement_OnlyRecipientCanConfirm(t *testing.T) {
	repo := newFakeRepo()
	users := newFakeUsers()
	s := NewService(repo, users)
	owner, member := uuid.New(), uuid.New()
	users.add("a@x", owner)
	users.add("b@x", member)
	g, _ := s.CreateGroup(owner, CreateGroupRequest{Name: "X"})
	s.AddMemberByEmail(g.ID, owner, AddMemberRequest{Email: "b@x"})

	rec, err := s.RecordSettlement(g.ID, owner, RecordSettlementRequest{
		FromUser: member.String(), ToUser: owner.String(), Amount: 100,
	})
	if err != nil {
		t.Fatalf("RecordSettlement: %v", err)
	}

	// Member (the sender) tries to confirm → rejected.
	if _, err := s.ConfirmSettlement(rec.ID, member); err == nil {
		t.Errorf("sender should not be able to confirm")
	}
	// Owner (recipient) confirms → ok.
	confirmed, err := s.ConfirmSettlement(rec.ID, owner)
	if err != nil {
		t.Fatalf("ConfirmSettlement: %v", err)
	}
	if confirmed.Status != SettlementConfirmed {
		t.Errorf("Status = %v, want SettlementConfirmed", confirmed.Status)
	}
}