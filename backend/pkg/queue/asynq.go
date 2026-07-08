// Package queue wraps Asynq for Redis-backed background job processing.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

// ─── Task type constants ───────────────────────────────────────────────────────

const (
	TypeOCR            = "ocr:process"
	TypeCategorize     = "categorize:expense"
	TypeTranscription  = "transcription:process"
	TypeWeeklySummary  = "summary:weekly"
	TypeMonthlySummary = "summary:monthly"
	TypeNotification   = "notification:send"
	TypeEmail          = "email:send"
	TypeBudgetCheck    = "budget:check"
	TypeRecurring      = "recurring:detect"
	TypeAnalytics      = "analytics:aggregate"
)

// ─── Payload structs ──────────────────────────────────────────────────────────

type OCRPayload          struct{ UserID, ImageURL string }
type CategorizePayload   struct{ UserID, Merchant, Description string }
type TranscriptionPayload struct{ UserID, AudioURL string }
type SummaryPayload      struct{ UserID, Type string }
type NotificationPayload struct{ UserID, Title, Message, NotifType string }
type EmailPayload        struct{ To, Subject, Body string }
type BudgetCheckPayload  struct{ UserID, Category string; Amount float64 }

// ─── Client (producer) ────────────────────────────────────────────────────────

// Client is the Asynq job producer.
type Client struct {
	asynq *asynq.Client
}

// NewClient creates a new Asynq client connected to Redis.
func NewClient(redisURL string) (*Client, error) {
	opt, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("queue: parse redis URL: %w", err)
	}
	return &Client{asynq: asynq.NewClient(opt)}, nil
}

// Close closes the Asynq client.
func (c *Client) Close() error { return c.asynq.Close() }

// EnqueueOCR enqueues an OCR processing job on the critical queue.
func (c *Client) EnqueueOCR(ctx context.Context, userID, imageURL string) error {
	return c.enqueue(ctx, TypeOCR, OCRPayload{userID, imageURL}, asynq.Queue("critical"), asynq.MaxRetry(3))
}

// EnqueueCategorize enqueues an AI categorisation job.
func (c *Client) EnqueueCategorize(ctx context.Context, userID, merchant, description string) error {
	return c.enqueue(ctx, TypeCategorize, CategorizePayload{userID, merchant, description})
}

// EnqueueTranscription enqueues a voice transcription job on the critical queue.
func (c *Client) EnqueueTranscription(ctx context.Context, userID, audioURL string) error {
	return c.enqueue(ctx, TypeTranscription, TranscriptionPayload{userID, audioURL}, asynq.Queue("critical"))
}

// EnqueueWeeklySummary enqueues a weekly summary generation job.
func (c *Client) EnqueueWeeklySummary(ctx context.Context, userID string) error {
	return c.enqueue(ctx, TypeWeeklySummary, SummaryPayload{userID, "weekly"})
}

// EnqueueMonthlySummary enqueues a monthly summary generation job.
func (c *Client) EnqueueMonthlySummary(ctx context.Context, userID string) error {
	return c.enqueue(ctx, TypeMonthlySummary, SummaryPayload{userID, "monthly"})
}

// EnqueueNotification enqueues an in-app / FCM notification job.
func (c *Client) EnqueueNotification(ctx context.Context, userID, title, message, notifType string) error {
	return c.enqueue(ctx, TypeNotification, NotificationPayload{userID, title, message, notifType})
}

// EnqueueEmail enqueues a transactional email job on the low queue.
func (c *Client) EnqueueEmail(ctx context.Context, to, subject, body string) error {
	return c.enqueue(ctx, TypeEmail, EmailPayload{to, subject, body}, asynq.Queue("low"))
}

// EnqueueBudgetCheck enqueues a budget overrun check after a new expense.
func (c *Client) EnqueueBudgetCheck(ctx context.Context, userID, category string, amount float64) error {
	return c.enqueue(ctx, TypeBudgetCheck, BudgetCheckPayload{userID, category, amount})
}

// EnqueueRecurringDetect enqueues a recurring expense detection job.
func (c *Client) EnqueueRecurringDetect(ctx context.Context, userID string) error {
	return c.enqueue(ctx, TypeRecurring, SummaryPayload{UserID: userID, Type: "recurring"})
}

// ─── generic enqueue helper ───────────────────────────────────────────────────

func (c *Client) enqueue(ctx context.Context, taskType string, payload interface{}, opts ...asynq.Option) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("queue: marshal payload: %w", err)
	}
	task := asynq.NewTask(taskType, data)
	if _, err := c.asynq.EnqueueContext(ctx, task, opts...); err != nil {
		return fmt.Errorf("queue: enqueue %s: %w", taskType, err)
	}
	return nil
}

// ─── Server (consumer) ────────────────────────────────────────────────────────

// ServerConfig holds Asynq server configuration.
type ServerConfig struct {
	RedisURL    string
	Concurrency int
}

// NewServer creates a new Asynq server.
func NewServer(cfg ServerConfig) (*asynq.Server, error) {
	opt, err := asynq.ParseRedisURI(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("queue: server parse redis URL: %w", err)
	}

	concurrency := cfg.Concurrency
	if concurrency == 0 {
		concurrency = 10
	}

	srv := asynq.NewServer(opt, asynq.Config{
		Concurrency: concurrency,
		Queues: map[string]int{
			"critical": 6,
			"default":  3,
			"low":      1,
		},
		RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
			return time.Duration(n*n) * time.Second
		},
	})
	return srv, nil
}
