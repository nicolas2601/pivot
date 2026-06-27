package accounts

type CreateRequest struct {
	Name           string  `json:"name" binding:"required,min=1,max=100"`
	Type           string  `json:"type" binding:"required,oneof=cash debit credit savings"`
	Currency       string  `json:"currency" binding:"required,len=3"`
	OpeningBalance int64   `json:"opening_balance"`
	Color          *string `json:"color,omitempty" binding:"omitempty,len=7"`
	Icon           *string `json:"icon,omitempty" binding:"omitempty,max=50"`
}

type UpdateRequest struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Color *string `json:"color,omitempty" binding:"omitempty,len=7"`
	Icon  *string `json:"icon,omitempty" binding:"omitempty,max=50"`
}

type ListResponse struct {
	Accounts []Account `json:"accounts"`
}