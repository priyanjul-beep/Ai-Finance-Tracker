// Package usecase – income business logic.
package usecase

import (
	"context"
	"time"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// IncomeUseCase implements interfaces.IncomeService.
type IncomeUseCase struct {
	incomes   interfaces.IncomeRepository
	auditLogs interfaces.AuditLogRepository
	cache     interfaces.CacheService
}

// NewIncome creates a new IncomeUseCase.
func NewIncome(
	incomes interfaces.IncomeRepository,
	auditLogs interfaces.AuditLogRepository,
	cache interfaces.CacheService,
) *IncomeUseCase {
	return &IncomeUseCase{
		incomes:   incomes,
		auditLogs: auditLogs,
		cache:     cache,
	}
}

// Create adds a new income record.
func (uc *IncomeUseCase) Create(ctx context.Context, userID string, req dto.CreateIncomeRequest) (*dto.IncomeDTO, error) {
	income := &domain.Income{
		UserID:        userID,
		Amount:        req.Amount,
		Currency:      "INR",
		Source:        req.Source,
		Category:      req.Category,
		Description:   req.Description,
		Notes:         req.Notes,
		Date:          req.Date,
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
		IsTaxable:     req.IsTaxable,
		TaxAmount:     req.TaxAmount,
	}

	if income.Currency == "" {
		income.Currency = "INR"
	}

	if err := uc.incomes.Create(ctx, income); err != nil {
		return nil, err
	}

	// Invalidate analytics cache
	_ = uc.cache.InvalidateUser(ctx, userID)

	// Audit log
	_ = uc.auditLogs.Create(ctx, &domain.AuditLog{
		UserID:   userID,
		Action:   "create",
		Entity:   "income",
		EntityID: income.ID,
	})

	return incomeToDTO(income), nil
}

// GetByID returns a single income entry belonging to the user.
func (uc *IncomeUseCase) GetByID(ctx context.Context, userID, incomeID string) (*dto.IncomeDTO, error) {
	income, err := uc.incomes.GetByID(ctx, incomeID)
	if err != nil {
		return nil, err
	}
	if income.UserID != userID {
		return nil, interfaces.ErrNotFound
	}
	return incomeToDTO(income), nil
}

// Update modifies an existing income record.
func (uc *IncomeUseCase) Update(ctx context.Context, userID, incomeID string, req dto.UpdateIncomeRequest) (*dto.IncomeDTO, error) {
	income, err := uc.incomes.GetByID(ctx, incomeID)
	if err != nil {
		return nil, err
	}
	if income.UserID != userID {
		return nil, interfaces.ErrNotFound
	}

	if req.Amount > 0 {
		income.Amount = req.Amount
	}
	if req.Source != "" {
		income.Source = req.Source
	}
	if req.Category != "" {
		income.Category = req.Category
	}
	if req.Description != "" {
		income.Description = req.Description
	}
	if req.Notes != "" {
		income.Notes = req.Notes
	}
	if !req.Date.IsZero() {
		income.Date = req.Date
	}
	if req.PaymentMethod != "" {
		income.PaymentMethod = domain.PaymentMethod(req.PaymentMethod)
	}
	income.IsTaxable = req.IsTaxable
	if req.TaxAmount >= 0 {
		income.TaxAmount = req.TaxAmount
	}

	if err := uc.incomes.Update(ctx, income); err != nil {
		return nil, err
	}

	_ = uc.cache.InvalidateUser(ctx, userID)
	return incomeToDTO(income), nil
}

// Delete removes an income record.
func (uc *IncomeUseCase) Delete(ctx context.Context, userID, incomeID string) error {
	income, err := uc.incomes.GetByID(ctx, incomeID)
	if err != nil {
		return err
	}
	if income.UserID != userID {
		return interfaces.ErrNotFound
	}

	if err := uc.incomes.Delete(ctx, incomeID); err != nil {
		return err
	}

	_ = uc.cache.InvalidateUser(ctx, userID)

	_ = uc.auditLogs.Create(ctx, &domain.AuditLog{
		UserID:   userID,
		Action:   "delete",
		Entity:   "income",
		EntityID: incomeID,
	})

	return nil
}

// List returns paginated income records for the given date range.
func (uc *IncomeUseCase) List(ctx context.Context, userID string, from, to time.Time, page, limit int) ([]*dto.IncomeDTO, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	var incomes []domain.Income
	var total int64
	var err error

	if !from.IsZero() && !to.IsZero() {
		// Date-range filtered (no pagination from repo layer – paginate in memory)
		incomes, err = uc.incomes.GetByDateRange(ctx, userID, from, to)
		if err != nil {
			return nil, 0, err
		}
		total = int64(len(incomes))
		// Apply offset/limit manually
		start := offset
		if start > len(incomes) {
			start = len(incomes)
		}
		end := start + limit
		if end > len(incomes) {
			end = len(incomes)
		}
		incomes = incomes[start:end]
	} else {
		incomes, total, err = uc.incomes.GetByUserID(ctx, userID, limit, offset)
		if err != nil {
			return nil, 0, err
		}
	}

	out := make([]*dto.IncomeDTO, len(incomes))
	for i, inc := range incomes {
		copy := inc // avoid loop var capture
		out[i] = incomeToDTO(&copy)
	}
	return out, total, nil
}

// GetMonthlyTotal returns the summed income for the given month.
func (uc *IncomeUseCase) GetMonthlyTotal(ctx context.Context, userID string, year, month int) (float64, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Second)
	return uc.incomes.TotalByDateRange(ctx, userID, from, to)
}

// ─── helper ─────────────────────────────────────────────────────────────────

func incomeToDTO(i *domain.Income) *dto.IncomeDTO {
	return &dto.IncomeDTO{
		ID:            i.ID,
		UserID:        i.UserID,
		Amount:        i.Amount,
		Currency:      i.Currency,
		Source:        i.Source,
		Category:      i.Category,
		Description:   i.Description,
		Notes:         i.Notes,
		Date:          i.Date,
		PaymentMethod: string(i.PaymentMethod),
		IsTaxable:     i.IsTaxable,
		TaxAmount:     i.TaxAmount,
		CreatedAt:     i.CreatedAt,
		UpdatedAt:     i.UpdatedAt,
	}
}
