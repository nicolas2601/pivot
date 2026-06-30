package transactions

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Create(tx *Transaction) error
	GetByID(id, userID uuid.UUID) (*Transaction, error)
	ListByUser(userID uuid.UUID, filter ListFilter) ([]Transaction, error)
	Update(tx *Transaction) error
	Delete(id, userID uuid.UUID) error

	// CreateTransfer atomically inserts a pair of transactions that represent
	// a movement of funds between two accounts. Both rows share the same
	// TransferPairID. The callback receives the (already-generated) pair id
	// so the implementation can set it on both rows.
	CreateTransfer(userID uuid.UUID, source, dest *Transaction) error

	// DeletePair soft-deletes both legs of a transfer identified by pairID.
	// Ownership is enforced by userID — only transactions belonging to that
	// user are touched.
	DeletePair(pairID, userID uuid.UUID) error

	// SumByCategory aggregates expenses by category between [from, to].
	SumByCategory(userID uuid.UUID, from, to time.Time) ([]CategorySum, error)
	// SumByAccount aggregates expenses by account between [from, to].
	SumByAccount(userID uuid.UUID, from, to time.Time) ([]AccountSum, error)
	// MonthlyTrend groups expenses by year-month between [from, to].
	MonthlyTrend(userID uuid.UUID, from, to time.Time) ([]MonthlyTotal, error)
	// AmountsByDay aggregates tx amounts by calendar date, filtered by tx
	// type. Returned map key is "YYYY-MM-DD". Used by /reports/summary +
	// /reports/cashflow so we don't parse string dates in app code.
	AmountsByDay(userID uuid.UUID, from, to time.Time, txType string) (map[string]int64, error)
	// MonthlyTrendAmountsByMonth — same as MonthlyTrend but parameterized on
	// type so we can build dual-line trend series (income + expense) in
	// the reports layer.
	MonthlyTrendAmountsByMonth(userID uuid.UUID, from, to time.Time, txType string) ([]MonthlyTotal, error)
}

// ListFilter holds optional filters for ListByUser. Zero values mean "no filter".
type ListFilter struct {
	AccountID  *uuid.UUID
	CategoryID *uuid.UUID
	Type       string
	From       *time.Time
	To         *time.Time
	Limit      int
}

type CategorySum struct {
	CategoryID *uuid.UUID `json:"category_id"`
	Total      int64      `json:"total"`
}

type AccountSum struct {
	AccountID uuid.UUID `json:"account_id"`
	Total     int64     `json:"total"`
}

type MonthlyTotal struct {
	Year  int   `json:"year"`
	Month int   `json:"month"`
	Total int64 `json:"total"`
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) Create(tx *Transaction) error {
	if tx.ID == uuid.Nil {
		tx.ID = uuid.New()
	}
	return r.db.Create(tx).Error
}

func (r *repo) GetByID(id, userID uuid.UUID) (*Transaction, error) {
	var t Transaction
	err := r.db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTransactionNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repo) ListByUser(userID uuid.UUID, f ListFilter) ([]Transaction, error) {
	var list []Transaction
	q := r.db.Where("user_id = ? AND deleted_at IS NULL", userID).Order("date DESC, created_at DESC")
	if f.AccountID != nil {
		q = q.Where("account_id = ?", *f.AccountID)
	}
	if f.CategoryID != nil {
		q = q.Where("category_id = ?", *f.CategoryID)
	}
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.From != nil {
		q = q.Where("date >= ?", *f.From)
	}
	if f.To != nil {
		q = q.Where("date <= ?", *f.To)
	}
	if f.Limit > 0 {
		q = q.Limit(f.Limit)
	}
	if err := q.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repo) Update(tx *Transaction) error {
	return r.db.Save(tx).Error
}

func (r *repo) Delete(id, userID uuid.UUID) error {
	return r.db.Model(&Transaction{}).
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *repo) CreateTransfer(userID uuid.UUID, source, dest *Transaction) error {
	if source.AccountID == dest.AccountID {
		return ErrAccountMismatch
	}
	if source.Currency != dest.Currency {
		return ErrCurrencyMismatch
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		pairID := uuid.New()
		source.TransferPairID = &pairID
		dest.TransferPairID = &pairID
		if source.ID == uuid.Nil {
			source.ID = uuid.New()
		}
		if dest.ID == uuid.Nil {
			dest.ID = uuid.New()
		}
		if err := tx.Create(source).Error; err != nil {
			return err
		}
		if err := tx.Create(dest).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *repo) DeletePair(pairID, userID uuid.UUID) error {
	return r.db.Model(&Transaction{}).
		Where("user_id = ? AND transfer_pair_id = ? AND deleted_at IS NULL", userID, pairID).
		Update("deleted_at", gorm.Expr("NOW()")).Error
}

func (r *repo) SumByCategory(userID uuid.UUID, from, to time.Time) ([]CategorySum, error) {
	var out []CategorySum
	err := r.db.Model(&Transaction{}).
		Select("category_id, SUM(amount) AS total").
		Where("user_id = ? AND deleted_at IS NULL AND type = ? AND date BETWEEN ? AND ?",
			userID, TypeExpense, from, to).
		Group("category_id").
		Find(&out).Error
	return out, err
}

// SumByCategoryAndWindow — same as SumByCategory but lets the caller
// also slice the date range so it stays a single query even when the
// caller needs sub-month precision (BudgetVsActual). Uses date_trunc to
// give the planner a sargable predicate that still hits the
// (user_id, date) index.
func (r *repo) SumByCategoryInDateRange(userID uuid.UUID, from, to time.Time) ([]CategorySum, error) {
	var out []CategorySum
	err := r.db.Model(&Transaction{}).
		Select("category_id, SUM(amount) AS total").
		Where("user_id = ? AND deleted_at IS NULL AND type = ? AND date >= ? AND date < ?",
			userID, TypeExpense, from, to).
		Group("category_id").
		Find(&out).Error
	return out, err
}

func (r *repo) SumByAccount(userID uuid.UUID, from, to time.Time) ([]AccountSum, error) {
	var out []AccountSum
	err := r.db.Model(&Transaction{}).
		Select("account_id, SUM(amount) AS total").
		Where("user_id = ? AND deleted_at IS NULL AND type = ? AND date BETWEEN ? AND ?",
			userID, TypeExpense, from, to).
		Group("account_id").
		Find(&out).Error
	return out, err
}

func (r *repo) MonthlyTrend(userID uuid.UUID, from, to time.Time) ([]MonthlyTotal, error) {
	return r.MonthlyTrendAmountsByMonth(userID, from, to, string(TypeExpense))
}

// MonthlyTrendAmountsByMonth is the same as MonthlyTrend but the tx
// type is passed in. Used by /reports/monthly-trend to render a dual-line
// sparkline (income + expense in two separate calls, joined in memory by
// the reports service).
func (r *repo) MonthlyTrendAmountsByMonth(userID uuid.UUID, from, to time.Time, txType string) ([]MonthlyTotal, error) {
	type row struct {
		MonthStart time.Time
		Total      int64
	}
	var rows []row
	err := r.db.Model(&Transaction{}).
		Select("date_trunc('month', date) AS month_start, SUM(amount) AS total").
		Where("user_id = ? AND deleted_at IS NULL AND type = ? AND date BETWEEN ? AND ?",
			userID, txType, from, to).
		Group("date_trunc('month', date)").
		Order("date_trunc('month', date) ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]MonthlyTotal, 0, len(rows))
	for _, r := range rows {
		y, m, _ := r.MonthStart.UTC().Date()
		out = append(out, MonthlyTotal{Year: y, Month: int(m), Total: r.Total})
	}
	return out, nil
}

// AmountsByDay returns a map of YYYY-MM-DD → total amount for the given
// tx type within [from, to]. Returned as map so the caller can read each
// day in O(1) without server-side grouping when merging income+expense.
func (r *repo) AmountsByDay(userID uuid.UUID, from, to time.Time, txType string) (map[string]int64, error) {
	type row struct {
		Date  time.Time
		Total int64
	}
	var rows []row
	err := r.db.Model(&Transaction{}).
		Select("date, SUM(amount) AS total").
		Where("user_id = ? AND deleted_at IS NULL AND type = ? AND date BETWEEN ? AND ?",
			userID, txType, from, to).
		Group("date").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[r.Date.UTC().Format("2006-01-02")] = r.Total
	}
	return out, nil
}