// Package queue – background task handlers (workers) registered with Asynq.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// HandlerDeps bundles all dependencies required by background workers.
// All fields are typed interfaces, populated by the DI container in main.go.
type HandlerDeps struct {
	AIProvider  interfaces.AIProvider
	EmailSvc    interfaces.EmailService
	NotifRepo   interfaces.NotificationRepository
	ExpenseRepo interfaces.ExpenseRepository
	BudgetRepo  interfaces.BudgetRepository
	UserRepo    interfaces.UserRepository
	Cache       interfaces.CacheService
}

// RegisterWorkers attaches all task handlers to the given ServeMux.
func RegisterWorkers(mux *asynq.ServeMux, deps HandlerDeps) {
	mux.HandleFunc(TypeOCR,                 newOCRHandler(deps))
	mux.HandleFunc(TypeCategorize,          newCategorizeHandler(deps))
	mux.HandleFunc(TypeTranscription,       newTranscriptionHandler(deps))
	mux.HandleFunc(TypeWeeklySummary,       newSummaryHandler(deps, "weekly"))
	mux.HandleFunc(TypeMonthlySummary,      newSummaryHandler(deps, "monthly"))
	mux.HandleFunc(TypeNotification,        newNotificationHandler(deps))
	mux.HandleFunc(TypeEmail,               newEmailHandler(deps))
	mux.HandleFunc(TypeBudgetCheck,         newBudgetCheckHandler(deps))
	mux.HandleFunc(TypeWelcomeNotif,        newWelcomeNotifHandler(deps))
	mux.HandleFunc(TypeBudgetWarningNotif,  newBudgetAlertHandler(deps))
	mux.HandleFunc(TypeBudgetExceededNotif, newBudgetAlertHandler(deps))
}

// ─── Welcome notification worker ─────────────────────────────────────────────

func newWelcomeNotifHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p WelcomeNotifPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("welcome_notif worker: unmarshal: %w", err)
		}
		log.Printf("[worker:welcome] user=%s email=%s", p.UserID, p.Email)

		// 1. Persist in-app notification
		notif := &domain.Notification{
			ID:       uuid.NewString(),
			UserID:   p.UserID,
			Title:    "Welcome to AI Finance Tracker 🎉",
			Message:  "Your account has been created successfully. Start adding expenses, create budgets, and track your financial goals.",
			Type:     domain.NotifWelcome,
			Priority: domain.PriorityLow,
		}
		if err := deps.NotifRepo.Create(ctx, notif); err != nil {
			log.Printf("[worker:welcome] error creating notification: %v", err)
			return fmt.Errorf("welcome_notif worker: create notification: %w", err)
		}
		log.Printf("[worker:welcome] notification created id=%s", notif.ID)

		// 2. Invalidate unread-count cache
		_ = deps.Cache.Delete(ctx, notifUnreadKey(p.UserID))

		// 3. Send welcome email
		if err := deps.EmailSvc.SendWelcomeHTML(ctx, p.Email, p.Name); err != nil {
			log.Printf("[worker:welcome] email send error (non-fatal): %v", err)
			// Non-fatal: notification was already created; email failure is logged
		}
		log.Printf("[worker:welcome] done user=%s", p.UserID)
		return nil
	}
}

// ─── Budget alert worker (warning + exceeded) ─────────────────────────────────

func newBudgetAlertHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p BudgetAlertPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("budget_alert worker: unmarshal: %w", err)
		}
		log.Printf("[worker:budget_alert] user=%s category=%s type=%s threshold=%.0f%%",
			p.UserID, p.Category, p.AlertType, p.Threshold)

		// Dedup check: never send same alert twice for same budget cycle
		dedupKey := fmt.Sprintf("notif_sent:%s:%s:%s:%.0f:%d:%d",
			p.UserID, p.BudgetID, p.AlertType, p.Threshold, p.Year, p.Month)
		var alreadySent bool
		if err := deps.Cache.Get(ctx, dedupKey, &alreadySent); err == nil && alreadySent {
			log.Printf("[worker:budget_alert] duplicate skipped key=%s", dedupKey)
			return nil // already sent
		}

		// 1. Build notification
		var (
			title    string
			message  string
			priority domain.NotificationPriority
			notifType domain.NotificationType
		)
		if p.AlertType == "exceeded" {
			title = "Budget Exceeded 🚨"
			message = fmt.Sprintf("Your %s budget has been exceeded. Review your recent expenses and adjust your spending.", p.Category)
			priority = domain.PriorityCritical
			notifType = domain.NotifBudgetExceeded
		} else {
			title = "Budget Alert ⚠️"
			message = fmt.Sprintf("You have used %.0f%% of your %s budget. Monitor your spending carefully.", p.Threshold, p.Category)
			priority = domain.PriorityHigh
			notifType = domain.NotifBudgetWarning
		}

		notif := &domain.Notification{
			ID:       uuid.NewString(),
			UserID:   p.UserID,
			Title:    title,
			Message:  message,
			Type:     notifType,
			Priority: priority,
		}
		if err := deps.NotifRepo.Create(ctx, notif); err != nil {
			return fmt.Errorf("budget_alert worker: create notification: %w", err)
		}

		// 2. Invalidate unread-count cache
		_ = deps.Cache.Delete(ctx, notifUnreadKey(p.UserID))

		// 3. Send email
		var emailErr error
		if p.AlertType == "exceeded" {
			overspent := p.Spent - p.BudgetAmount
			if overspent < 0 {
				overspent = 0
			}
			emailErr = deps.EmailSvc.SendBudgetExceededHTML(ctx, p.Email, p.Name, p.Category, p.BudgetAmount, p.Spent, overspent)
		} else {
			emailErr = deps.EmailSvc.SendBudgetWarningHTML(ctx, p.Email, p.Name, p.Category,
				p.BudgetAmount, p.Spent, p.Remaining, p.Threshold,
				p.Month, p.Year, p.DaysLeft)
		}
		if emailErr != nil {
			log.Printf("[worker:budget_alert] email send error (non-fatal): %v", emailErr)
		}

		// 4. Mark dedup key – expires after 35 days (covers full month)
		_ = deps.Cache.Set(ctx, dedupKey, true, 35*24*time.Hour)

		log.Printf("[worker:budget_alert] done user=%s category=%s type=%s notif_id=%s",
			p.UserID, p.Category, p.AlertType, notif.ID)
		return nil
	}
}

// ─── Budget-check handler (triggered after each expense) ────────────────────

func newBudgetCheckHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p BudgetCheckPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("budget_check worker: unmarshal: %w", err)
		}
		log.Printf("[worker:budget_check] user=%s category=%s amount=%.2f", p.UserID, p.Category, p.Amount)

		// Fetch budget for user/category
		budget, err := deps.BudgetRepo.GetByUserAndCategory(ctx, p.UserID, p.Category)
		if err != nil || budget == nil || !budget.IsActive {
			return nil // no budget configured for this category
		}

		// Compute current month spending
		now := time.Now()
		month := budget.Month
		year := budget.Year
		if month == 0 {
			month = int(now.Month())
		}
		if year == 0 {
			year = now.Year()
		}
		from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0).Add(-time.Second)

		categorySpends, err := deps.ExpenseRepo.SumByCategory(ctx, p.UserID, from, to)
		if err != nil {
			return fmt.Errorf("budget_check worker: sum by category: %w", err)
		}
		var spent float64
		for _, cs := range categorySpends {
			if cs.Category == p.Category {
				spent = cs.Amount
				break
			}
		}

		if budget.Amount <= 0 || spent <= 0 {
			return nil
		}

		pct := (spent / budget.Amount) * 100
		remaining := budget.Amount - spent
		daysLeft := int(to.Sub(now).Hours() / 24)
		if daysLeft < 0 {
			daysLeft = 0
		}

		// Fetch user details for email
		user, err := deps.UserRepo.GetByID(ctx, p.UserID)
		if err != nil {
			return fmt.Errorf("budget_check worker: get user: %w", err)
		}

		// Helper: enqueue budget alert job if threshold not already sent
		enqueueAlert := func(alertType string, threshold float64) {
			dedupKey := fmt.Sprintf("notif_sent:%s:%s:%s:%.0f:%d:%d",
				p.UserID, budget.ID, alertType, threshold, year, month)
			var alreadySent bool
			if err := deps.Cache.Get(ctx, dedupKey, &alreadySent); err == nil && alreadySent {
				return // dedup
			}
			payload := BudgetAlertPayload{
				UserID:       p.UserID,
				Email:        user.Email,
				Name:         user.Name,
				BudgetID:     budget.ID,
				Category:     p.Category,
				BudgetAmount: budget.Amount,
				Spent:        spent,
				Remaining:    remaining,
				Threshold:    threshold,
				AlertType:    alertType,
				Month:        month,
				Year:         year,
				DaysLeft:     daysLeft,
			}
			// Use a new client to avoid self-referencing – enqueue inline
			notif := &domain.Notification{
				ID:       uuid.NewString(),
				UserID:   p.UserID,
				Title:    buildAlertTitle(alertType, threshold),
				Message:  buildAlertMessage(alertType, threshold, p.Category),
				Type:     alertTypeToNotifType(alertType),
				Priority: alertTypeToPriority(alertType),
			}
			if createErr := deps.NotifRepo.Create(ctx, notif); createErr != nil {
				log.Printf("[worker:budget_check] create notif error: %v", createErr)
				return
			}
			_ = deps.Cache.Delete(ctx, notifUnreadKey(p.UserID))

			var emailErr error
			if alertType == "exceeded" {
				overspent := spent - budget.Amount
				if overspent < 0 {
					overspent = 0
				}
				emailErr = deps.EmailSvc.SendBudgetExceededHTML(ctx, user.Email, user.Name, p.Category, budget.Amount, spent, overspent)
			} else {
				emailErr = deps.EmailSvc.SendBudgetWarningHTML(ctx, user.Email, user.Name, p.Category,
					budget.Amount, spent, remaining, threshold, month, year, daysLeft)
			}
			if emailErr != nil {
				log.Printf("[worker:budget_check] email error (non-fatal): %v", emailErr)
			}
			_ = deps.Cache.Set(ctx, dedupKey, true, 35*24*time.Hour)
			log.Printf("[worker:budget_check] alert sent user=%s category=%s type=%s threshold=%.0f",
				p.UserID, p.Category, alertType, threshold)
			_ = payload // suppress unused warning — payload data is inlined above
		}

		// Check thresholds (order matters: exceeded first, then 90%, then custom AlertAt)
		if pct >= 100 {
			enqueueAlert("exceeded", 100)
		} else if pct >= 90 {
			enqueueAlert("warning", 90)
		} else if budget.AlertAt > 0 && budget.AlertAt < 90 && pct >= budget.AlertAt {
			enqueueAlert("warning", budget.AlertAt)
		}

		return nil
	}
}

// ─── OCR handler ─────────────────────────────────────────────────────────────

func newOCRHandler(_ HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p OCRPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("ocr worker: unmarshal: %w", err)
		}
		log.Printf("[worker:ocr] processing image for user=%s url=%s", p.UserID, p.ImageURL)
		// TODO: call OCR service, extract text, call AI to parse expense, persist result
		return nil
	}
}

// ─── Categorize handler ───────────────────────────────────────────────────────

func newCategorizeHandler(_ HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p CategorizePayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("categorize worker: unmarshal: %w", err)
		}
		log.Printf("[worker:categorize] merchant=%s description=%s", p.Merchant, p.Description)
		return nil
	}
}

// ─── Transcription handler ────────────────────────────────────────────────────

func newTranscriptionHandler(_ HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p TranscriptionPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("transcription worker: unmarshal: %w", err)
		}
		log.Printf("[worker:transcription] audio=%s user=%s", p.AudioURL, p.UserID)
		return nil
	}
}

// ─── Summary handler ──────────────────────────────────────────────────────────

func newSummaryHandler(_ HandlerDeps, summaryType string) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p SummaryPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("summary worker (%s): unmarshal: %w", summaryType, err)
		}
		log.Printf("[worker:summary:%s] user=%s", summaryType, p.UserID)
		return nil
	}
}

// ─── Notification handler ─────────────────────────────────────────────────────

func newNotificationHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p NotificationPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("notification worker: unmarshal: %w", err)
		}
		log.Printf("[worker:notification] type=%s user=%s", p.NotifType, p.UserID)
		notif := &domain.Notification{
			ID:       uuid.NewString(),
			UserID:   p.UserID,
			Title:    p.Title,
			Message:  p.Message,
			Type:     domain.NotificationType(p.NotifType),
			Priority: domain.PriorityLow,
		}
		if err := deps.NotifRepo.Create(ctx, notif); err != nil {
			return fmt.Errorf("notification worker: create: %w", err)
		}
		_ = deps.Cache.Delete(ctx, notifUnreadKey(p.UserID))
		return nil
	}
}

// ─── Email handler ────────────────────────────────────────────────────────────

func newEmailHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p EmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("email worker: unmarshal: %w", err)
		}
		log.Printf("[worker:email] to=%s subject=%s", p.To, p.Subject)
		return deps.EmailSvc.SendBudgetAlert(ctx, p.To, p.Body)
	}
}

// ─── Shared helpers ───────────────────────────────────────────────────────────

// notifUnreadKey returns the Redis cache key for a user's unread notification count.
func notifUnreadKey(userID string) string {
	return fmt.Sprintf("notif_unread:%s", userID)
}

func buildAlertTitle(alertType string, threshold float64) string {
	if alertType == "exceeded" {
		return "Budget Exceeded 🚨"
	}
	if threshold >= 90 {
		return "Budget Alert ⚠️"
	}
	return fmt.Sprintf("Budget Alert: %.0f%% Used ⚠️", threshold)
}

func buildAlertMessage(alertType string, threshold float64, category string) string {
	if alertType == "exceeded" {
		return fmt.Sprintf("Your %s monthly budget has been exceeded. Review your recent expenses and adjust your spending.", category)
	}
	return fmt.Sprintf("You have already used %.0f%% of your %s budget. Please monitor your spending.", threshold, category)
}

func alertTypeToNotifType(alertType string) domain.NotificationType {
	if alertType == "exceeded" {
		return domain.NotifBudgetExceeded
	}
	return domain.NotifBudgetWarning
}

func alertTypeToPriority(alertType string) domain.NotificationPriority {
	if alertType == "exceeded" {
		return domain.PriorityCritical
	}
	return domain.PriorityHigh
}

// Ensure dto import is used (for future summary workers).
var _ *dto.MonthlyReportDTO
