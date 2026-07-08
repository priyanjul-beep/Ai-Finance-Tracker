// Package usecase – financial goal business logic.
package usecase

import (
	"context"
	"time"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// GoalUseCase implements goal management.
type GoalUseCase struct {
	goals     interfaces.GoalRepository
	auditLogs interfaces.AuditLogRepository
	cache     interfaces.CacheService
}

// NewGoal creates a new GoalUseCase.
func NewGoal(
	goals interfaces.GoalRepository,
	auditLogs interfaces.AuditLogRepository,
	cache interfaces.CacheService,
) *GoalUseCase {
	return &GoalUseCase{goals: goals, auditLogs: auditLogs, cache: cache}
}

// Create adds a financial goal.
func (uc *GoalUseCase) Create(ctx context.Context, userID string, req dto.CreateGoalRequest) (*dto.GoalDTO, error) {
	goal := &domain.Goal{
		UserID:       userID,
		Name:         req.Name,
		Description:  req.Description,
		TargetAmount: req.TargetAmount,
		Currency:     "INR",
		Category:     req.Category,
		TargetDate:   req.TargetDate,
		Priority:     req.Priority,
		Status:       "active",
	}
	if goal.Priority == 0 {
		goal.Priority = 3
	}

	if err := uc.goals.Create(ctx, goal); err != nil {
		return nil, err
	}

	_ = uc.auditLogs.Create(ctx, &domain.AuditLog{
		UserID:   userID,
		Action:   "create",
		Entity:   "goal",
		EntityID: goal.ID,
	})

	return goalToDTO(goal), nil
}

// GetByID returns a goal belonging to the user.
func (uc *GoalUseCase) GetByID(ctx context.Context, userID, goalID string) (*dto.GoalDTO, error) {
	goal, err := uc.goals.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if goal.UserID != userID {
		return nil, interfaces.ErrNotFound
	}
	return goalToDTO(goal), nil
}

// Update modifies a goal.
func (uc *GoalUseCase) Update(ctx context.Context, userID, goalID string, req dto.UpdateGoalRequest) (*dto.GoalDTO, error) {
	goal, err := uc.goals.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if goal.UserID != userID {
		return nil, interfaces.ErrNotFound
	}

	if req.Name != "" {
		goal.Name = req.Name
	}
	if req.Description != "" {
		goal.Description = req.Description
	}
	if req.TargetAmount > 0 {
		goal.TargetAmount = req.TargetAmount
	}
	if !req.TargetDate.IsZero() {
		goal.TargetDate = req.TargetDate
	}
	if req.Priority > 0 {
		goal.Priority = req.Priority
	}
	if req.Status != "" {
		goal.Status = req.Status
	}

	// Auto-complete when target is reached
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.Status = "completed"
	}

	if err := uc.goals.Update(ctx, goal); err != nil {
		return nil, err
	}

	return goalToDTO(goal), nil
}

// Contribute adds an amount to the current saved amount of the goal.
func (uc *GoalUseCase) Contribute(ctx context.Context, userID, goalID string, amount float64) (*dto.GoalDTO, error) {
	goal, err := uc.goals.GetByID(ctx, goalID)
	if err != nil {
		return nil, err
	}
	if goal.UserID != userID {
		return nil, interfaces.ErrNotFound
	}

	goal.CurrentAmount += amount
	if goal.CurrentAmount > goal.TargetAmount {
		goal.CurrentAmount = goal.TargetAmount
	}
	if goal.CurrentAmount >= goal.TargetAmount {
		goal.Status = "completed"
	}

	if err := uc.goals.Update(ctx, goal); err != nil {
		return nil, err
	}

	return goalToDTO(goal), nil
}

// Delete removes a goal.
func (uc *GoalUseCase) Delete(ctx context.Context, userID, goalID string) error {
	goal, err := uc.goals.GetByID(ctx, goalID)
	if err != nil {
		return err
	}
	if goal.UserID != userID {
		return interfaces.ErrNotFound
	}
	return uc.goals.Delete(ctx, goalID)
}

// List returns all goals for the user, optionally filtered by status.
func (uc *GoalUseCase) List(ctx context.Context, userID string, status string) ([]*dto.GoalDTO, error) {
	goals, err := uc.goals.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*dto.GoalDTO, 0, len(goals))
	for _, g := range goals {
		if status != "" && g.Status != status {
			continue
		}
		copy := g
		out = append(out, goalToDTO(&copy))
	}
	return out, nil
}

// ─── helper ──────────────────────────────────────────────────────────────────

func goalToDTO(g *domain.Goal) *dto.GoalDTO {
	var progressPct float64
	if g.TargetAmount > 0 {
		progressPct = (g.CurrentAmount / g.TargetAmount) * 100
	}

	// Days remaining
	daysRemaining := int(time.Until(g.TargetDate).Hours() / 24)
	if daysRemaining < 0 {
		daysRemaining = 0
	}

	// Monthly savings needed
	var monthlySavingsNeeded float64
	remaining := g.TargetAmount - g.CurrentAmount
	if remaining > 0 && daysRemaining > 0 {
		months := float64(daysRemaining) / 30.0
		if months > 0 {
			monthlySavingsNeeded = remaining / months
		}
	}

	return &dto.GoalDTO{
		ID:                   g.ID,
		UserID:               g.UserID,
		Name:                 g.Name,
		Description:          g.Description,
		TargetAmount:         g.TargetAmount,
		CurrentAmount:        g.CurrentAmount,
		Currency:             g.Currency,
		Category:             g.Category,
		TargetDate:           g.TargetDate,
		Priority:             g.Priority,
		Status:               g.Status,
		ProgressPercent:      progressPct,
		DaysRemaining:        daysRemaining,
		MonthlySavingsNeeded: monthlySavingsNeeded,
		CreatedAt:            g.CreatedAt,
		UpdatedAt:            g.UpdatedAt,
	}
}
