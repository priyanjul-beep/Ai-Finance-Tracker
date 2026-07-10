// Package usecase – budget business logic.
package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// BudgetUseCase implements interfaces.BudgetService.
type BudgetUseCase struct {
	budgets   interfaces.BudgetRepository
	expenses  interfaces.ExpenseRepository
	queue     interfaces.QueueService
	auditLogs interfaces.AuditLogRepository
	notifs    interfaces.NotificationRepository
}

// NewBudget creates a new BudgetUseCase.
func NewBudget(
	budgets interfaces.BudgetRepository,
	expenses interfaces.ExpenseRepository,
	queue interfaces.QueueService,
	auditLogs interfaces.AuditLogRepository,
	notifs interfaces.NotificationRepository,
) *BudgetUseCase {
	return &BudgetUseCase{
		budgets:   budgets,
		expenses:  expenses,
		queue:     queue,
		auditLogs: auditLogs,
		notifs:    notifs,
	}
}

// Create adds a new budget.
func (uc *BudgetUseCase) Create(ctx context.Context, userID string, req dto.CreateBudgetRequest) (*dto.BudgetStatusDTO, error) {
	budget := &domain.Budget{
		UserID:      userID,
		Category:    domain.ExpenseCategory(req.Category),
		Amount:      req.Amount,
		Currency:    "INR",
		Period:      req.Period,
		Month:       req.Month,
		Year:        req.Year,
		AlertAt:     req.AlertAt,
		Description: req.Description,
		IsActive:    true,
	}
	if budget.Period == "" {
		budget.Period = "monthly"
	}
	if budget.AlertAt == 0 {
		budget.AlertAt = 80
	}

	if err := uc.budgets.Create(ctx, budget); err != nil {
		return nil, err
	}

	return uc.enrichWithSpending(ctx, userID, budget)
}

// GetByID returns a budget with current spending status.
func (uc *BudgetUseCase) GetByID(ctx context.Context, userID, budgetID string) (*dto.BudgetStatusDTO, error) {
	budget, err := uc.budgets.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}
	if budget.UserID != userID {
		return nil, interfaces.ErrNotFound
	}
	return uc.enrichWithSpending(ctx, userID, budget)
}

// Update modifies a budget.
func (uc *BudgetUseCase) Update(ctx context.Context, userID, budgetID string, req dto.UpdateBudgetRequest) (*dto.BudgetStatusDTO, error) {
	budget, err := uc.budgets.GetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}
	if budget.UserID != userID {
		return nil, interfaces.ErrNotFound
	}

	if req.Amount > 0 {
		budget.Amount = req.Amount
	}
	if req.AlertAt > 0 {
		budget.AlertAt = req.AlertAt
	}
	if req.Description != "" {
		budget.Description = req.Description
	}
	if req.Period != "" {
		budget.Period = req.Period
	}

	if err := uc.budgets.Update(ctx, budget); err != nil {
		return nil, err
	}
	return uc.enrichWithSpending(ctx, userID, budget)
}

// Delete removes a budget.
func (uc *BudgetUseCase) Delete(ctx context.Context, userID, budgetID string) error {
	budget, err := uc.budgets.GetByID(ctx, budgetID)
	if err != nil {
		return err
	}
	if budget.UserID != userID {
		return interfaces.ErrNotFound
	}
	return uc.budgets.Delete(ctx, budgetID)
}

// List returns all active budgets for the user in the given year/month, enriched with spending.
func (uc *BudgetUseCase) List(ctx context.Context, userID string, year, month int, category string) ([]*dto.BudgetStatusDTO, error) {
	budgets, err := uc.budgets.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*dto.BudgetStatusDTO, 0, len(budgets))
	for _, b := range budgets {
		if !b.IsActive {
			continue
		}
		if category != "" && !strings.EqualFold(string(b.Category), category) {
			continue
		}
		bCopy := b
		enriched, err := uc.enrichWithSpending(ctx, userID, &bCopy)
		if err != nil {
			continue
		}
		out = append(out, enriched)
	}

	// Check alerts for any at/over threshold budgets (deduped per day)
	if uc.notifs != nil {
		go uc.checkAlerts(context.Background(), userID, out)
	}

	return out, nil
}

// checkAlerts fires budget_warning notifications for budgets at/over threshold.
func (uc *BudgetUseCase) checkAlerts(ctx context.Context, userID string, budgets []*dto.BudgetStatusDTO) {
	for _, b := range budgets {
		if b.Percent < b.AlertAt {
			continue
		}
		var title, message string
		if b.Spent >= b.Amount {
			title = fmt.Sprintf("Over budget: %s", b.Category)
			message = fmt.Sprintf("You've spent ₹%.0f of your ₹%.0f %s budget (%.0f%% used).",
				b.Spent, b.Amount, b.Category, b.Percent)
		} else {
			title = fmt.Sprintf("Budget alert: %s at %.0f%%", b.Category, b.Percent)
			message = fmt.Sprintf("You've used %.0f%% of your ₹%.0f %s budget (₹%.0f spent).",
				b.Percent, b.Amount, b.Category, b.Spent)
		}
		if exists, _ := uc.notifs.ExistsToday(ctx, userID, "budget_warning", title); exists {
			continue
		}
		_ = uc.notifs.Create(ctx, &domain.Notification{
			ID:      uuid.NewString(),
			UserID:  userID,
			Title:   title,
			Message: message,
			Type:    "budget_warning",
		})
	}
}

// ─── helpers ─────────────────────────────────────────────────────────────────

func (uc *BudgetUseCase) enrichWithSpending(ctx context.Context, userID string, b *domain.Budget) (*dto.BudgetStatusDTO, error) {
	// Calculate month window
	now := time.Now()
	year := b.Year
	month := b.Month
	if month == 0 {
		month = int(now.Month())
	}
	if year == 0 {
		year = now.Year()
	}
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Second)

	// Get all-category sums and pick the one matching this budget
	categorySpends, err := uc.expenses.SumByCategory(ctx, userID, from, to)
	var spent float64
	if err == nil {
		for _, cs := range categorySpends {
			if strings.EqualFold(cs.Category, string(b.Category)) {
				spent = cs.Amount
				break
			}
		}
	}

	remaining := b.Amount - spent
	if remaining < 0 {
		remaining = 0
	}

	var pct float64
	if b.Amount > 0 {
		pct = (spent / b.Amount) * 100
	}

	status := "on-track"
	switch {
	case spent >= b.Amount:
		status = "over-budget"
	case pct >= b.AlertAt:
		status = "warning"
	}

	return &dto.BudgetStatusDTO{
		BudgetDTO: dto.BudgetDTO{
			ID:          b.ID,
			UserID:      b.UserID,
			Category:    string(b.Category),
			Amount:      b.Amount,
			Currency:    b.Currency,
			Period:      b.Period,
			Month:       b.Month,
			Year:        b.Year,
			AlertAt:     b.AlertAt,
			Description: b.Description,
			IsActive:    b.IsActive,
			CreatedAt:   b.CreatedAt,
			UpdatedAt:   b.UpdatedAt,
		},
		Spent:     spent,
		Remaining: remaining,
		Percent:   pct,
		Status:    status,
	}, nil
}
