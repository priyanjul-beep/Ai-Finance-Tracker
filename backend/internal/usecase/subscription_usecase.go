// Package usecase – subscription business logic.
package usecase

import (
	"context"
	"time"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// SubscriptionUseCase implements subscription management.
type SubscriptionUseCase struct {
	subs      interfaces.SubscriptionRepository
	auditLogs interfaces.AuditLogRepository
	cache     interfaces.CacheService
}

// NewSubscription creates a new SubscriptionUseCase.
func NewSubscription(
	subs interfaces.SubscriptionRepository,
	auditLogs interfaces.AuditLogRepository,
	cache interfaces.CacheService,
) *SubscriptionUseCase {
	return &SubscriptionUseCase{subs: subs, auditLogs: auditLogs, cache: cache}
}

// Create adds a subscription.
func (uc *SubscriptionUseCase) Create(ctx context.Context, userID string, req dto.CreateSubscriptionRequest) (*dto.SubscriptionDTO, error) {
	sub := &domain.Subscription{
		UserID:          userID,
		Name:            req.Name,
		Amount:          req.Amount,
		Currency:        "INR",
		BillingCycle:    req.BillingCycle,
		NextBillingDate: req.NextBillingDate,
		Category:        domain.ExpenseCategory(req.Category),
		PaymentMethod:   domain.PaymentMethod(req.PaymentMethod),
		Notes:           req.Notes,
		IsActive:        true,
	}
	if sub.BillingCycle == "" {
		sub.BillingCycle = "monthly"
	}

	if err := uc.subs.Create(ctx, sub); err != nil {
		return nil, err
	}

	_ = uc.cache.InvalidateUser(ctx, userID)
	return subToDTO(sub), nil
}

// GetByID returns a subscription owned by the user.
func (uc *SubscriptionUseCase) GetByID(ctx context.Context, userID, subID string) (*dto.SubscriptionDTO, error) {
	sub, err := uc.subs.GetByID(ctx, subID)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, interfaces.ErrNotFound
	}
	return subToDTO(sub), nil
}

// Update modifies a subscription.
func (uc *SubscriptionUseCase) Update(ctx context.Context, userID, subID string, req dto.UpdateSubscriptionRequest) (*dto.SubscriptionDTO, error) {
	sub, err := uc.subs.GetByID(ctx, subID)
	if err != nil {
		return nil, err
	}
	if sub.UserID != userID {
		return nil, interfaces.ErrNotFound
	}

	if req.Name != "" {
		sub.Name = req.Name
	}
	if req.Amount > 0 {
		sub.Amount = req.Amount
	}
	if req.BillingCycle != "" {
		sub.BillingCycle = req.BillingCycle
	}
	if !req.NextBillingDate.IsZero() {
		sub.NextBillingDate = req.NextBillingDate
	}
	if req.Category != "" {
		sub.Category = domain.ExpenseCategory(req.Category)
	}
	if req.PaymentMethod != "" {
		sub.PaymentMethod = domain.PaymentMethod(req.PaymentMethod)
	}
	if req.Notes != "" {
		sub.Notes = req.Notes
	}

	if err := uc.subs.Update(ctx, sub); err != nil {
		return nil, err
	}

	_ = uc.cache.InvalidateUser(ctx, userID)
	return subToDTO(sub), nil
}

// Delete removes a subscription.
func (uc *SubscriptionUseCase) Delete(ctx context.Context, userID, subID string) error {
	sub, err := uc.subs.GetByID(ctx, subID)
	if err != nil {
		return err
	}
	if sub.UserID != userID {
		return interfaces.ErrNotFound
	}
	_ = uc.cache.InvalidateUser(ctx, userID)
	return uc.subs.Delete(ctx, subID)
}

// List returns all subscriptions for the user.
func (uc *SubscriptionUseCase) List(ctx context.Context, userID string, activeOnly bool) ([]*dto.SubscriptionDTO, error) {
	subs, err := uc.subs.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*dto.SubscriptionDTO, 0, len(subs))
	for _, s := range subs {
		if activeOnly && !s.IsActive {
			continue
		}
		copy := s
		out = append(out, subToDTO(&copy))
	}
	return out, nil
}

// GetUpcoming returns subscriptions due within the next N days.
func (uc *SubscriptionUseCase) GetUpcoming(ctx context.Context, userID string, days int) ([]*dto.SubscriptionDTO, error) {
	if days <= 0 {
		days = 7
	}
	dueBy := time.Now().AddDate(0, 0, days)

	all, err := uc.subs.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	out := make([]*dto.SubscriptionDTO, 0)
	for _, s := range all {
		if s.IsActive && !s.NextBillingDate.After(dueBy) {
			out = append(out, subToDTO(&s))
		}
	}
	return out, nil
}

// ─── helper ──────────────────────────────────────────────────────────────────

func subToDTO(s *domain.Subscription) *dto.SubscriptionDTO {
	return &dto.SubscriptionDTO{
		ID:              s.ID,
		UserID:          s.UserID,
		Name:            s.Name,
		Amount:          s.Amount,
		Currency:        s.Currency,
		BillingCycle:    s.BillingCycle,
		NextBillingDate: s.NextBillingDate,
		Category:        string(s.Category),
		PaymentMethod:   string(s.PaymentMethod),
		Notes:           s.Notes,
		IsActive:        s.IsActive,
		CreatedAt:       s.CreatedAt,
		UpdatedAt:       s.UpdatedAt,
	}
}
