package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// ExpenseUseCase implements interfaces.ExpenseService.
type ExpenseUseCase struct {
	expenses     interfaces.ExpenseRepository
	tags         interfaces.TagRepository
	auditLogs    interfaces.AuditLogRepository
	ai           interfaces.AIProvider
	cache        interfaces.CacheService
	queue        interfaces.QueueService
	merchantRepo interfaces.MerchantMappingRepository
}

// NewExpense creates a new ExpenseUseCase.
func NewExpense(
	expenses interfaces.ExpenseRepository,
	tags interfaces.TagRepository,
	auditLogs interfaces.AuditLogRepository,
	ai interfaces.AIProvider,
	cache interfaces.CacheService,
	queue interfaces.QueueService,
	merchantRepo interfaces.MerchantMappingRepository,
) *ExpenseUseCase {
	return &ExpenseUseCase{
		expenses: expenses, tags: tags, auditLogs: auditLogs,
		ai: ai, cache: cache, queue: queue, merchantRepo: merchantRepo,
	}
}

// Create persists a new expense, performs duplicate detection, and enqueues
// background jobs for categorisation and budget checking.
func (uc *ExpenseUseCase) Create(ctx context.Context, userID string, req dto.CreateExpenseRequest) (*dto.ExpenseDTO, error) {
	// Resolve category: prefer request value, fall back to AI
	category := req.Category
	if category == "" || category == "unknown" {
		cached, _ := uc.cache.GetMerchantCategory(ctx, req.Merchant)
		if cached != "" {
			category = cached
		} else {
			// Enqueue async categorisation; use "others" meanwhile
			_ = uc.queue.EnqueueCategorize(ctx, userID, req.Merchant, req.Description)
			category = "others"
		}
	}

	expense := &domain.Expense{
		ID:            uuid.NewString(),
		UserID:        userID,
		Amount:        req.Amount,
		Currency:      "INR",
		Category:      domain.ExpenseCategory(category),
		Merchant:      req.Merchant,
		Description:   req.Description,
		Notes:         req.Notes,
		Date:          req.Date,
		ExpenseType:   domain.ExpenseType(req.ExpenseType),
		PaymentMethod: domain.PaymentMethod(req.PaymentMethod),
	}
	if expense.ExpenseType == "" {
		expense.ExpenseType = domain.ExpenseTypeSpend
	}

	// Duplicate detection
	dups, _ := uc.expenses.FindDuplicates(ctx, userID, req.Amount, req.Merchant)
	if len(dups) > 0 {
		expense.IsDuplicate = true
		dupID := dups[0].ID
		expense.DuplicateOf = &dupID
	}

	if err := uc.expenses.Create(ctx, expense); err != nil {
		return nil, fmt.Errorf("expense: create: %w", err)
	}

	// Attach tags
	for _, tagID := range req.Tags {
		_ = uc.tags.AddToExpense(ctx, tagID, expense.ID)
	}

	// Enqueue budget overrun check
	_ = uc.queue.EnqueueBudgetCheck(ctx, userID, category, req.Amount)

	// Invalidate dashboard cache
	_ = uc.cache.InvalidateUser(ctx, userID)

	// Write audit log
	_ = uc.auditLogs.Create(ctx, &domain.AuditLog{
		ID:       uuid.NewString(),
		UserID:   userID,
		Action:   "CREATE",
		Entity:   "expense",
		EntityID: expense.ID,
	})

	return mapExpense(expense), nil
}

// GetByID returns an expense, enforcing ownership.
func (uc *ExpenseUseCase) GetByID(ctx context.Context, userID, id string) (*dto.ExpenseDTO, error) {
	expense, err := uc.expenses.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if expense.UserID != userID {
		return nil, fmt.Errorf("expense not found")
	}
	return mapExpense(expense), nil
}

// List returns a paginated list of expenses for the user.
func (uc *ExpenseUseCase) List(ctx context.Context, userID string, p dto.PaginationParams) (*dto.PaginatedResponse, error) {
	offset := (p.Page - 1) * p.Limit
	rows, total, err := uc.expenses.GetByUserID(ctx, userID, p.Limit, offset)
	if err != nil {
		return nil, err
	}

	var items []dto.ExpenseDTO
	for i := range rows {
		items = append(items, *mapExpense(&rows[i]))
	}

	totalPages := total / int64(p.Limit)
	if total%int64(p.Limit) != 0 {
		totalPages++
	}

	return &dto.PaginatedResponse{
		Data: items,
		Pagination: dto.Pagination{
			Page: p.Page, Limit: p.Limit,
			Total: total, TotalPages: totalPages,
			HasNext: int64(p.Page) < totalPages,
			HasPrev: p.Page > 1,
		},
	}, nil
}

// Update modifies an existing expense.
func (uc *ExpenseUseCase) Update(ctx context.Context, userID, id string, req dto.UpdateExpenseRequest) (*dto.ExpenseDTO, error) {
	expense, err := uc.expenses.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if expense.UserID != userID {
		return nil, fmt.Errorf("expense not found")
	}

	if req.Amount > 0 {
		expense.Amount = req.Amount
	}
	if req.Category != "" {
		expense.Category = domain.ExpenseCategory(req.Category)
	}
	if req.Merchant != "" {
		expense.Merchant = req.Merchant
	}
	if req.Description != "" {
		expense.Description = req.Description
	}
	if req.Notes != "" {
		expense.Notes = req.Notes
	}
	if !req.Date.IsZero() {
		expense.Date = req.Date
	}
	if req.PaymentMethod != "" {
		expense.PaymentMethod = domain.PaymentMethod(req.PaymentMethod)
	}
	if req.IsFavorite != nil {
		expense.IsFavorite = *req.IsFavorite
	}

	if err := uc.expenses.Update(ctx, expense); err != nil {
		return nil, err
	}

	_ = uc.cache.InvalidateUser(ctx, userID)
	return mapExpense(expense), nil
}

// Delete soft-deletes an expense.
func (uc *ExpenseUseCase) Delete(ctx context.Context, userID, id string) error {
	expense, err := uc.expenses.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if expense.UserID != userID {
		return fmt.Errorf("expense not found")
	}
	_ = uc.cache.InvalidateUser(ctx, userID)
	return uc.expenses.Delete(ctx, id)
}

// ParseFromText calls the AI provider to extract expense fields from text.
func (uc *ExpenseUseCase) ParseFromText(ctx context.Context, _ string, req dto.AIExpenseParseRequest) (*dto.AIExpenseParseResponse, error) {
	return uc.ai.ParseExpense(ctx, req.Text, req.ImageURL)
}

// ParseFromVoice transcribes audio and extracts expense data via Gemini
// multimodal processing. Results are cached by SHA-256 of the raw audio bytes
// so identical recordings hit Redis instead of the AI API.
func (uc *ExpenseUseCase) ParseFromVoice(ctx context.Context, _ string, audioData []byte, mimeType string) (*dto.AIVoiceParseResponse, error) {
	// Build cache key from audio hash
	sum := sha256.Sum256(audioData)
	hash := fmt.Sprintf("%x", sum)

	// Serve from cache if available
	var cached dto.AIVoiceParseResponse
	if err := uc.cache.GetVoiceCache(ctx, hash, &cached); err == nil {
		cached.Cached = true
		return &cached, nil
	}

	// Parse via Gemini multimodal
	result, err := uc.ai.ParseExpenseFromAudio(ctx, audioData, mimeType)
	if err != nil {
		return nil, fmt.Errorf("voice parse: %w", err)
	}

	// Persist to cache (best-effort; don't fail on Redis errors)
	_ = uc.cache.SetVoiceCache(ctx, hash, result)

	return result, nil
}

// Search converts a natural-language query to SQL and returns matching expenses.
func (uc *ExpenseUseCase) Search(ctx context.Context, userID, query string) ([]dto.ExpenseDTO, error) {
	whereClause, err := uc.ai.NLToSQLFilter(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("expense search: ai filter: %w", err)
	}

	// Execute parameterised query with the AI-generated WHERE clause
	_ = whereClause // TODO: wire to repository.SearchByFilter(ctx, userID, whereClause)
	return []dto.ExpenseDTO{}, nil
}

// GetDuplicates returns potential duplicates of a given expense.
func (uc *ExpenseUseCase) GetDuplicates(ctx context.Context, userID, id string) ([]dto.ExpenseDTO, error) {
	expense, err := uc.expenses.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if expense.UserID != userID {
		return nil, fmt.Errorf("expense not found")
	}

	dups, err := uc.expenses.FindDuplicates(ctx, userID, expense.Amount, expense.Merchant)
	if err != nil {
		return nil, err
	}

	var result []dto.ExpenseDTO
	for i := range dups {
		if dups[i].ID != id {
			result = append(result, *mapExpense(&dups[i]))
		}
	}
	return result, nil
}

// ─── mapping ──────────────────────────────────────────────────────────────────

func mapExpense(e *domain.Expense) *dto.ExpenseDTO {
	d := &dto.ExpenseDTO{
		ID:            e.ID,
		Amount:        e.Amount,
		Currency:      e.Currency,
		Category:      string(e.Category),
		Merchant:      e.Merchant,
		Description:   e.Description,
		Notes:         e.Notes,
		Date:          e.Date,
		ExpenseType:   string(e.ExpenseType),
		PaymentMethod: string(e.PaymentMethod),
		ImageURL:      e.ImageURL,
		IsFavorite:    e.IsFavorite,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
	for _, t := range e.Tags {
		d.Tags = append(d.Tags, dto.TagDTO{ID: t.ID, Name: t.Name, Color: t.Color})
	}
	return d
}

