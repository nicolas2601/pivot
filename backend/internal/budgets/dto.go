package budgets

type CreateRequestDTO struct {
	CategoryID string  `json:"category_id" binding:"required,uuid"`
	Amount     int64   `json:"amount" binding:"required,gt=0"`
	Period     string  `json:"period" binding:"required,oneof=monthly yearly"`
	StartDate  string  `json:"start_date" binding:"required"`
	EndDate    *string `json:"end_date,omitempty" binding:"omitempty"`
}

type UpdateRequestDTO struct {
	Amount       *int64  `json:"amount,omitempty" binding:"omitempty,gt=0"`
	Period       *string `json:"period,omitempty" binding:"omitempty,oneof=monthly yearly"`
	StartDate    *string `json:"start_date,omitempty"`
	EndDate      *string `json:"end_date,omitempty"`
	ClearEndDate bool    `json:"clear_end_date,omitempty"`
}

type ListResponse struct {
	Budgets []Budget `json:"budgets"`
}

func (d CreateRequestDTO) ToServiceCreate() CreateRequest {
	return CreateRequest{
		CategoryID: d.CategoryID,
		Amount:     d.Amount,
		Period:     d.Period,
		StartDate:  d.StartDate,
		EndDate:    d.EndDate,
	}
}

func (d UpdateRequestDTO) ToServiceUpdate() UpdateRequest {
	return UpdateRequest{
		Amount:    d.Amount,
		Period:    d.Period,
		StartDate: d.StartDate,
		EndDate:   d.EndDate,
		ClearEnd:  d.ClearEndDate,
	}
}