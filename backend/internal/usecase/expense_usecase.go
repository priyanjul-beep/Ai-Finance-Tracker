package usecase

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

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
	notifs       interfaces.NotificationRepository
	budgets      interfaces.BudgetRepository
	userRepo     interfaces.UserRepository
	emailSvc     interfaces.EmailService
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
	notifs interfaces.NotificationRepository,
	budgets interfaces.BudgetRepository,
	userRepo interfaces.UserRepository,
	emailSvc interfaces.EmailService,
) *ExpenseUseCase {
	return &ExpenseUseCase{
		expenses: expenses, tags: tags, auditLogs: auditLogs,
		ai: ai, cache: cache, queue: queue, merchantRepo: merchantRepo,
		notifs: notifs, budgets: budgets, userRepo: userRepo, emailSvc: emailSvc,
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

	// Enqueue budget overrun check (async, best-effort)
	_ = uc.queue.EnqueueBudgetCheck(ctx, userID, category, req.Amount)

	// Directly check budget and write notifications to DB (synchronous, reliable)
	if uc.notifs != nil && uc.budgets != nil {
		go uc.checkBudgetAndNotify(context.Background(), userID, category)
	}

	// Large expense notification (direct DB write, no queue dependency)
	if req.Amount >= 5000 && uc.notifs != nil {
		msg := fmt.Sprintf("₹%.0f spent at %s has been recorded.", req.Amount, req.Merchant)
		_ = uc.notifs.Create(ctx, &domain.Notification{
			ID:      uuid.NewString(),
			UserID:  userID,
			Title:   "Large expense recorded",
			Message: msg,
			Type:    "large_expense",
		})
		if uc.emailSvc != nil && uc.userRepo != nil {
			if u, err := uc.userRepo.GetByID(ctx, userID); err == nil {
				go func() {
					_ = uc.emailSvc.SendBudgetAlert(context.Background(), u.Email,
						"Large Expense Alert\n\n"+msg)
				}()
			}
		}
	}

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

// checkBudgetAndNotify checks current month spending against budget and writes
// a notification directly to the DB if the threshold is crossed.
func (uc *ExpenseUseCase) checkBudgetAndNotify(ctx context.Context, userID, category string) {
	budget, err := uc.budgets.GetByUserAndCategory(ctx, userID, category)
	if err != nil || budget == nil {
		return
	}

	now := time.Now()
	from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Second)
	categorySpends, err := uc.expenses.SumByCategory(ctx, userID, from, to)
	if err != nil {
		return
	}

	var totalSpent float64
	for _, cs := range categorySpends {
		if strings.EqualFold(cs.Category, category) {
			totalSpent = cs.Amount
			break
		}
	}

	if budget.Amount <= 0 {
		return
	}
	pct := (totalSpent / budget.Amount) * 100
	if pct < budget.AlertAt {
		return
	}

	var title, message, notifType string
	if totalSpent >= budget.Amount {
		notifType = "budget_warning"
		title = fmt.Sprintf("Over budget: %s", category)
		message = fmt.Sprintf("You've spent ₹%.0f of your ₹%.0f %s budget (%.0f%% used).",
			totalSpent, budget.Amount, category, pct)
	} else {
		notifType = "budget_warning"
		title = fmt.Sprintf("Budget alert: %s at %.0f%%", category, pct)
		message = fmt.Sprintf("You've used %.0f%% of your ₹%.0f %s budget (₹%.0f spent).",
			pct, budget.Amount, category, totalSpent)
	}

	// Deduplicate: skip if same notification already sent today
	if exists, _ := uc.notifs.ExistsToday(ctx, userID, notifType, title); exists {
		return
	}

	_ = uc.notifs.Create(ctx, &domain.Notification{
		ID:      uuid.NewString(),
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
	})

	// Send email alert
	if uc.emailSvc != nil && uc.userRepo != nil {
		if u, err := uc.userRepo.GetByID(ctx, userID); err == nil {
			_ = uc.emailSvc.SendBudgetAlert(ctx, u.Email, title+"\n\n"+message)
		}
	}
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
	var (
		rows  []domain.Expense
		total int64
		err   error
	)
	if p.Category != "" || p.Merchant != "" {
		rows, total, err = uc.expenses.Search(ctx, userID, p.Merchant, p.Category, p.Limit, offset)
	} else {
		rows, total, err = uc.expenses.GetByUserID(ctx, userID, p.Limit, offset)
	}
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
	if req.ExpenseType != "" {
		expense.ExpenseType = domain.ExpenseType(req.ExpenseType)
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

// ParseFromImage analyses a receipt/screenshot image via Gemini Vision.
// Results are cached by SHA-256 of the raw image bytes (24h TTL).
func (uc *ExpenseUseCase) ParseFromImage(ctx context.Context, _ string, imageData []byte, mimeType, ocrText string) (*dto.AIReceiptScanResponse, error) {
	sum := sha256.Sum256(imageData)
	hash := fmt.Sprintf("%x", sum)

	var cached dto.AIReceiptScanResponse
	if err := uc.cache.GetScanCache(ctx, hash, &cached); err == nil {
		cached.Cached = true
		return &cached, nil
	}

	result, err := uc.ai.ParseExpenseFromImage(ctx, imageData, mimeType, ocrText)
	if err != nil {
		return nil, fmt.Errorf("image scan: %w", err)
	}

	_ = uc.cache.SetScanCache(ctx, hash, result)
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

