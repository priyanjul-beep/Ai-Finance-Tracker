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
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// HandlerDeps bundles the dependencies required by every worker.
type HandlerDeps struct {
	AIProvider  interface{}
	EmailSvc    interfaces.EmailService
	NotifRepo   interfaces.NotificationRepository
	ExpenseRepo interfaces.ExpenseRepository
	BudgetRepo  interfaces.BudgetRepository
}

// RegisterWorkers attaches all task handlers to the given ServeMux.
func RegisterWorkers(mux *asynq.ServeMux, deps HandlerDeps) {
	mux.HandleFunc(TypeOCR,            newOCRHandler(deps))
	mux.HandleFunc(TypeCategorize,     newCategorizeHandler(deps))
	mux.HandleFunc(TypeTranscription,  newTranscriptionHandler(deps))
	mux.HandleFunc(TypeWeeklySummary,  newSummaryHandler(deps, "weekly"))
	mux.HandleFunc(TypeMonthlySummary, newSummaryHandler(deps, "monthly"))
	mux.HandleFunc(TypeNotification,   newNotificationHandler(deps))
	mux.HandleFunc(TypeEmail,          newEmailHandler(deps))
	mux.HandleFunc(TypeBudgetCheck,    newBudgetCheckHandler(deps))
}

// ─── OCR handler ──────────────────────────────────────────────────────────────

func newOCRHandler(_ HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p OCRPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("ocr worker: unmarshal: %w", err)
		}
		log.Printf("[worker:ocr] processing image for user=%s url=%s", p.UserID, p.ImageURL)
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
// Persists an in-app notification row in the database.

func newNotificationHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p NotificationPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("notification worker: unmarshal: %w", err)
		}
		if deps.NotifRepo == nil {
			return nil
		}
		n := &domain.Notification{
			ID:      uuid.NewString(),
			UserID:  p.UserID,
			Title:   p.Title,
			Message: p.Message,
			Type:    p.NotifType,
			IsRead:  false,
		}
		if err := deps.NotifRepo.Create(ctx, n); err != nil {
			return fmt.Errorf("notification worker: create: %w", err)
		}
		log.Printf("[worker:notification] created type=%s for user=%s", p.NotifType, p.UserID)
		return nil
	}
}

// ─── Email handler ────────────────────────────────────────────────────────────
// Sends a transactional email via the SMTP service.

func newEmailHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p EmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("email worker: unmarshal: %w", err)
		}
		if deps.EmailSvc == nil {
			return nil
		}
		if err := deps.EmailSvc.SendBudgetAlert(ctx, p.To, p.Body); err != nil {
			return fmt.Errorf("email worker: send to %s: %w", p.To, err)
		}
		log.Printf("[worker:email] sent to=%s subject=%s", p.To, p.Subject)
		return nil
	}
}

// ─── Budget-check handler ─────────────────────────────────────────────────────
// Checks if adding `amount` to `category` spending pushes it over the budget
// threshold and, if so, creates a warning notification + sends an email.

func newBudgetCheckHandler(deps HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p BudgetCheckPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("budget_check worker: unmarshal: %w", err)
		}
		if deps.BudgetRepo == nil || deps.NotifRepo == nil || deps.ExpenseRepo == nil {
			return nil
		}

		// Look up the budget for this user+category
		budget, err := deps.BudgetRepo.GetByUserAndCategory(ctx, p.UserID, p.Category)
		if err != nil || budget == nil {
			return nil // no budget set for this category – nothing to check
		}

		// Get total spending in the current month for this category
		now := time.Now()
		from := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0).Add(-time.Second)
		categorySpends, err := deps.ExpenseRepo.SumByCategory(ctx, p.UserID, from, to)
		if err != nil {
			return nil
		}

		var totalSpent float64
		for _, cs := range categorySpends {
			if cs.Category == p.Category {
				totalSpent = cs.Amount
				break
			}
		}

		pct := 0.0
		if budget.Amount > 0 {
			pct = (totalSpent / budget.Amount) * 100
		}

		// Only notify when crossing the alert threshold OR going over budget
		if pct < budget.AlertAt {
			return nil
		}

		var title, message, notifType string
		if totalSpent >= budget.Amount {
			notifType = "budget_warning"
			title = fmt.Sprintf("Over budget: %s", p.Category)
			message = fmt.Sprintf(
				"You've spent ₹%.0f of your ₹%.0f %s budget (%.0f%%).",
				totalSpent, budget.Amount, p.Category, pct,
			)
		} else {
			notifType = "budget_warning"
			title = fmt.Sprintf("Budget alert: %s at %.0f%%", p.Category, pct)
			message = fmt.Sprintf(
				"You've used %.0f%% of your ₹%.0f %s budget (₹%.0f spent).",
				pct, budget.Amount, p.Category, totalSpent,
			)
		}

		// Persist in-app notification
		n := &domain.Notification{
			ID:      uuid.NewString(),
			UserID:  p.UserID,
			Title:   title,
			Message: message,
			Type:    notifType,
			IsRead:  false,
		}
		_ = deps.NotifRepo.Create(ctx, n)
		log.Printf("[worker:budget_check] alert for user=%s category=%s pct=%.0f%%", p.UserID, p.Category, pct)
		return nil
	}
}
