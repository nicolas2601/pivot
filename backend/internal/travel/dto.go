package travel

type CreateGroupDTO struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
	Currency    string  `json:"currency" binding:"omitempty,len=3"`
}

type UpdateGroupDTO struct {
	Name        *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

type GroupResponse struct {
	Group TravelGroup `json:"group"`
}

type GroupsResponse struct {
	Groups []TravelGroup `json:"groups"`
}

type AddMemberDTO struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"omitempty,oneof=owner member"`
}

type MembersResponse struct {
	Members []TravelGroupMember `json:"members"`
}

type ExpenseShareInputDTO struct {
	UserID string `json:"user_id" binding:"required,uuid"`
	Amount int64  `json:"amount"`
}

type AddExpenseDTO struct {
	PaidBy      string                  `json:"paid_by" binding:"required,uuid"`
	Amount      int64                   `json:"amount" binding:"required,gt=0"`
	Currency    string                  `json:"currency" binding:"omitempty,len=3"`
	Description string                  `json:"description" binding:"required,min=1,max=255"`
	SplitMethod string                  `json:"split_method" binding:"omitempty,oneof=equal exact percentage"`
	Date        string                  `json:"date"`
	Shares      []ExpenseShareInputDTO  `json:"shares"`
}

type ExpenseResponse struct {
	Expense TravelExpense       `json:"expense"`
	Shares  []TravelExpenseShare `json:"shares"`
}

type ExpensesResponse struct {
	Expenses []TravelExpense `json:"expenses"`
}

type SettlementSuggestionResponse struct {
	Suggestions []SettlementSuggestion `json:"suggestions"`
}

type RecordSettlementDTO struct {
	FromUser string `json:"from_user" binding:"required,uuid"`
	ToUser   string `json:"to_user" binding:"required,uuid"`
	Amount   int64  `json:"amount" binding:"required,gt=0"`
}

type SettlementsResponse struct {
	Settlements []TravelSettlement `json:"settlements"`
}

func (d CreateGroupDTO) ToServiceCreate() CreateGroupRequest {
	return CreateGroupRequest{
		Name:        d.Name,
		Description: d.Description,
		Currency:    d.Currency,
	}
}

func (d UpdateGroupDTO) ToServiceUpdate() UpdateGroupRequest {
	return UpdateGroupRequest{
		Name:        d.Name,
		Description: d.Description,
	}
}

func (d AddExpenseDTO) ToServiceAdd() (AddExpenseRequest, error) {
	shares := make([]ShareInput, 0, len(d.Shares))
	for _, s := range d.Shares {
		shares = append(shares, ShareInput{UserID: s.UserID, Amount: s.Amount})
	}
	return AddExpenseRequest{
		PaidBy:      d.PaidBy,
		Amount:      d.Amount,
		Currency:    d.Currency,
		Description: d.Description,
		SplitMethod: d.SplitMethod,
		Date:        d.Date,
		Shares:      shares,
	}, nil
}

func (d RecordSettlementDTO) ToServiceRecord() RecordSettlementRequest {
	return RecordSettlementRequest{
		FromUser: d.FromUser,
		ToUser:   d.ToUser,
		Amount:   d.Amount,
	}
}