// Package queue – background task handlers (workers) registered with Asynq.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

// HandlerDeps bundles the dependencies required by every worker.
// Populated by the DI container in main.go and passed to RegisterWorkers.
type HandlerDeps struct {
	// These will be satisfied by concrete service implementations at wire-up.
	// Using interface{} here keeps this package decoupled; production code
	// asserts to the appropriate interface.
	AIProvider   interface{}
	EmailSvc     interface{}
	NotifRepo    interface{}
	ExpenseRepo  interface{}
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
		// TODO: call AIProvider.CategorizeExpense and update the expense row
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
		// TODO: call Whisper/speech-recognition service, then AI to parse expense
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
		// TODO: collect expense/income data, call AIProvider.GenerateSummary,
		//       store result, notify user via email + in-app notification
		return nil
	}
}

// ─── Notification handler ─────────────────────────────────────────────────────

func newNotificationHandler(_ HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p NotificationPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("notification worker: unmarshal: %w", err)
		}
		log.Printf("[worker:notification] type=%s user=%s", p.NotifType, p.UserID)
		// TODO: persist in notifications table + send FCM push if token available
		return nil
	}
}

// ─── Email handler ────────────────────────────────────────────────────────────

func newEmailHandler(_ HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p EmailPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("email worker: unmarshal: %w", err)
		}
		log.Printf("[worker:email] to=%s subject=%s", p.To, p.Subject)
		// TODO: call EmailService.Send
		return nil
	}
}

// ─── Budget-check handler ─────────────────────────────────────────────────────

func newBudgetCheckHandler(_ HandlerDeps) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p BudgetCheckPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			return fmt.Errorf("budget_check worker: unmarshal: %w", err)
		}
		log.Printf("[worker:budget_check] user=%s category=%s amount=%.2f", p.UserID, p.Category, p.Amount)
		// TODO: query budget, compare spend vs limit, send warning notification
		return nil
	}
}
