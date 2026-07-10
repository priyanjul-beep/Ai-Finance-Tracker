// Package interfaces defines every port (repository + service) that separates
// the domain from infrastructure.  All concrete implementations must satisfy
// these contracts.
package interfaces

import (
	"context"
	"errors"
	"time"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
)

// Sentinel errors returned by repositories and use-cases.
var (
	ErrNotFound     = errors.New("record not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrConflict     = errors.New("already exists")
)

// ╔══════════════════════════════════════════════════════════════════════════════
// ║  Repository Interfaces
// ╚══════════════════════════════════════════════════════════════════════════════

// UserRepository abstracts persistence for the User aggregate.
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id string) error
}

// SessionRepository manages JWT session persistence for token revocation.
type SessionRepository interface {
	Create(ctx context.Context, session *domain.Session) error
	GetByAccessToken(ctx context.Context, token string) (*domain.Session, error)
	GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	DeleteByID(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

// ExpenseRepository abstracts persistence for the Expense aggregate.
type ExpenseRepository interface {
	Create(ctx context.Context, expense *domain.Expense) error
	GetByID(ctx context.Context, id string) (*domain.Expense, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Expense, int64, error)
	Search(ctx context.Context, userID, merchant, category string, limit, offset int) ([]domain.Expense, int64, error)
	Update(ctx context.Context, expense *domain.Expense) error
	Delete(ctx context.Context, id string) error
	GetByDateRange(ctx context.Context, userID string, from, to time.Time) ([]domain.Expense, error)
	GetByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error)
	GetByMerchant(ctx context.Context, userID, merchant string) ([]domain.Expense, error)
	FindDuplicates(ctx context.Context, userID string, amount float64, merchant string) ([]domain.Expense, error)
	SumByCategory(ctx context.Context, userID string, from, to time.Time) ([]dto.CategorySpend, error)
	SumByMerchant(ctx context.Context, userID string, from, to time.Time) ([]dto.MerchantSpend, error)
	TotalByDateRange(ctx context.Context, userID string, from, to time.Time) (float64, error)
}

// IncomeRepository abstracts persistence for the Income aggregate.
type IncomeRepository interface {
	Create(ctx context.Context, income *domain.Income) error
	GetByID(ctx context.Context, id string) (*domain.Income, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Income, int64, error)
	Search(ctx context.Context, userID, source, category string, limit, offset int) ([]domain.Income, int64, error)
	Update(ctx context.Context, income *domain.Income) error
	Delete(ctx context.Context, id string) error
	GetByDateRange(ctx context.Context, userID string, from, to time.Time) ([]domain.Income, error)
	TotalByDateRange(ctx context.Context, userID string, from, to time.Time) (float64, error)
}

// BudgetRepository abstracts persistence for the Budget aggregate.
type BudgetRepository interface {
	Create(ctx context.Context, budget *domain.Budget) error
	GetByID(ctx context.Context, id string) (*domain.Budget, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Budget, error)
	GetByUserAndCategory(ctx context.Context, userID, category string) (*domain.Budget, error)
	Update(ctx context.Context, budget *domain.Budget) error
	Delete(ctx context.Context, id string) error
}

// SubscriptionRepository abstracts persistence for Subscription.
type SubscriptionRepository interface {
	Create(ctx context.Context, sub *domain.Subscription) error
	GetByID(ctx context.Context, id string) (*domain.Subscription, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Subscription, error)
	Update(ctx context.Context, sub *domain.Subscription) error
	Delete(ctx context.Context, id string) error
	GetDueSubscriptions(ctx context.Context) ([]domain.Subscription, error)
}

// GoalRepository abstracts persistence for Goal.
type GoalRepository interface {
	Create(ctx context.Context, goal *domain.Goal) error
	GetByID(ctx context.Context, id string) (*domain.Goal, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Goal, error)
	Update(ctx context.Context, goal *domain.Goal) error
	Delete(ctx context.Context, id string) error
}

// TagRepository abstracts persistence for Tag.
type TagRepository interface {
	Create(ctx context.Context, tag *domain.Tag) error
	GetByID(ctx context.Context, id string) (*domain.Tag, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.Tag, error)
	Update(ctx context.Context, tag *domain.Tag) error
	Delete(ctx context.Context, id string) error
	AddToExpense(ctx context.Context, tagID, expenseID string) error
	RemoveFromExpense(ctx context.Context, tagID, expenseID string) error
}

// NotificationRepository abstracts persistence for Notification.
type NotificationRepository interface {
	Create(ctx context.Context, n *domain.Notification) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, int64, error)
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, id string) error
}

// AuditLogRepository writes append-only audit records.
type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.AuditLog, int64, error)
}

// MerchantMappingRepository is used for merchant→category look-up.
type MerchantMappingRepository interface {
	Create(ctx context.Context, m *domain.MerchantMapping) error
	GetByMerchant(ctx context.Context, merchant string) (*domain.MerchantMapping, error)
	Update(ctx context.Context, m *domain.MerchantMapping) error
	List(ctx context.Context) ([]domain.MerchantMapping, error)
}

// RecurringExpenseRepository manages detected recurring patterns.
type RecurringExpenseRepository interface {
	Create(ctx context.Context, e *domain.RecurringExpense) error
	GetByID(ctx context.Context, id string) (*domain.RecurringExpense, error)
	GetByUserID(ctx context.Context, userID string) ([]domain.RecurringExpense, error)
	Update(ctx context.Context, e *domain.RecurringExpense) error
	Delete(ctx context.Context, id string) error
}

// FinancialHealthScoreRepository stores the latest computed score per user.
type FinancialHealthScoreRepository interface {
	Upsert(ctx context.Context, score *domain.FinancialHealthScore) error
	GetByUserID(ctx context.Context, userID string) (*domain.FinancialHealthScore, error)
}

// ╔══════════════════════════════════════════════════════════════════════════════
// ║  Application Service Interfaces
// ╚══════════════════════════════════════════════════════════════════════════════

// AuthService covers the full authentication lifecycle.
type AuthService interface {
	Signup(ctx context.Context, req dto.SignupRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthResponse, error)
	VerifyEmail(ctx context.Context, token string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, req dto.PasswordResetConfirm) error
	ChangePassword(ctx context.Context, userID string, req dto.ChangePasswordRequest) error
	Logout(ctx context.Context, userID string) error
	GoogleOAuth(ctx context.Context, code string) (*dto.AuthResponse, error)
}

// UserService manages the user profile.
type UserService interface {
	GetProfile(ctx context.Context, userID string) (*dto.UserDTO, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateUserRequest) (*dto.UserDTO, error)
	DeleteAccount(ctx context.Context, userID string) error
}

// ExpenseService manages expense operations.
type ExpenseService interface {
	Create(ctx context.Context, userID string, req dto.CreateExpenseRequest) (*dto.ExpenseDTO, error)
	GetByID(ctx context.Context, userID, id string) (*dto.ExpenseDTO, error)
	List(ctx context.Context, userID string, p dto.PaginationParams) (*dto.PaginatedResponse, error)
	Update(ctx context.Context, userID, id string, req dto.UpdateExpenseRequest) (*dto.ExpenseDTO, error)
	Delete(ctx context.Context, userID, id string) error
	ParseFromText(ctx context.Context, userID string, req dto.AIExpenseParseRequest) (*dto.AIExpenseParseResponse, error)
	ParseFromVoice(ctx context.Context, userID string, audioData []byte, mimeType string) (*dto.AIVoiceParseResponse, error)
	ParseFromImage(ctx context.Context, userID string, imageData []byte, mimeType, ocrText string) (*dto.AIReceiptScanResponse, error)
	Search(ctx context.Context, userID, query string) ([]dto.ExpenseDTO, error)
	GetDuplicates(ctx context.Context, userID, id string) ([]dto.ExpenseDTO, error)
}

// IncomeService manages income operations.
type IncomeService interface {
	Create(ctx context.Context, userID string, req dto.CreateIncomeRequest) (*dto.IncomeDTO, error)
	GetByID(ctx context.Context, userID, id string) (*dto.IncomeDTO, error)
	List(ctx context.Context, userID string, from, to time.Time, source, category string, page, limit int) ([]*dto.IncomeDTO, int64, error)
	Update(ctx context.Context, userID, id string, req dto.UpdateIncomeRequest) (*dto.IncomeDTO, error)
	Delete(ctx context.Context, userID, id string) error
	GetMonthlyTotal(ctx context.Context, userID string, year, month int) (float64, error)
}

// BudgetService manages budgets and warns on overruns.
type BudgetService interface {
	Create(ctx context.Context, userID string, req dto.CreateBudgetRequest) (*dto.BudgetStatusDTO, error)
	GetByID(ctx context.Context, userID, id string) (*dto.BudgetStatusDTO, error)
	List(ctx context.Context, userID string, year, month int, category string) ([]*dto.BudgetStatusDTO, error)
	Update(ctx context.Context, userID, id string, req dto.UpdateBudgetRequest) (*dto.BudgetStatusDTO, error)
	Delete(ctx context.Context, userID, id string) error
}

// AnalyticsService computes dashboard and report data.
type AnalyticsService interface {
	GetDashboard(ctx context.Context, userID string) (*dto.DashboardDTO, error)
	GetMonthlyReport(ctx context.Context, userID string, month, year int) (*dto.MonthlyReportDTO, error)
	GetYearlyReport(ctx context.Context, userID string, year int) (map[string]interface{}, error)
	GetPredictions(ctx context.Context, userID string) (*dto.PredictionData, error)
	GetInsights(ctx context.Context, userID string) ([]string, error)
	GetFinancialHealthScore(ctx context.Context, userID string) (*domain.FinancialHealthScore, error)
}

// ╔══════════════════════════════════════════════════════════════════════════════
// ║  Infrastructure Service Interfaces
// ╚══════════════════════════════════════════════════════════════════════════════

// AIProvider is the pluggable AI backend (Gemini, OpenAI, Claude, etc.).
type AIProvider interface {
	ParseExpense(ctx context.Context, text, imageURL string) (*dto.AIExpenseParseResponse, error)
	ParseExpenseFromAudio(ctx context.Context, audioData []byte, mimeType string) (*dto.AIVoiceParseResponse, error)
	ParseExpenseFromImage(ctx context.Context, imageData []byte, mimeType, ocrText string) (*dto.AIReceiptScanResponse, error)
	CategorizeExpense(ctx context.Context, merchant, description string) (string, error)
	GenerateSummary(ctx context.Context, data string, summaryType string) (string, error)
	GenerateInsights(ctx context.Context, data map[string]interface{}) ([]string, error)
	NLToSQLFilter(ctx context.Context, query, userID string) (string, error)
	PredictExpenses(ctx context.Context, data map[string]interface{}) (*dto.PredictionData, error)
	CalcHealthScore(ctx context.Context, data map[string]interface{}) (float64, error)
}

// CacheService wraps Redis operations used across the application.
type CacheService interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	IncrementRateLimit(ctx context.Context, key string, ttl time.Duration) (int64, error)
	GetMerchantCategory(ctx context.Context, merchant string) (string, error)
	SetMerchantCategory(ctx context.Context, merchant, category string) error
	GetDashboard(ctx context.Context, userID string, dest interface{}) error
	SetDashboard(ctx context.Context, userID string, data interface{}) error
	InvalidateUser(ctx context.Context, userID string) error
	// Voice-parse cache: keyed by SHA-256 hash of raw audio bytes (24h TTL).
	GetVoiceCache(ctx context.Context, hash string, dest interface{}) error
	SetVoiceCache(ctx context.Context, hash string, data interface{}) error
	// Receipt-scan cache: keyed by SHA-256 hash of raw image bytes (24h TTL).
	GetScanCache(ctx context.Context, hash string, dest interface{}) error
	SetScanCache(ctx context.Context, hash string, data interface{}) error
}

// QueueService enqueues background jobs via Asynq.
type QueueService interface {
	EnqueueOCR(ctx context.Context, userID, imageURL string) error
	EnqueueCategorize(ctx context.Context, userID, merchant, description string) error
	EnqueueTranscription(ctx context.Context, userID, audioURL string) error
	EnqueueWeeklySummary(ctx context.Context, userID string) error
	EnqueueMonthlySummary(ctx context.Context, userID string) error
	EnqueueNotification(ctx context.Context, userID, title, message, notifType string) error
	EnqueueEmail(ctx context.Context, to, subject, body string) error
	EnqueueBudgetCheck(ctx context.Context, userID, category string, amount float64) error
	EnqueueRecurringDetect(ctx context.Context, userID string) error
}

// StorageService abstracts local and cloud file storage.
type StorageService interface {
	Upload(ctx context.Context, filename string, data []byte, mimeType string) (string, error)
	Delete(ctx context.Context, fileURL string) error
	GetURL(ctx context.Context, path string) string
}

// EmailService sends transactional emails.
type EmailService interface {
	SendVerification(ctx context.Context, to, link string) error
	SendPasswordReset(ctx context.Context, to, link string) error
	SendMonthlyReport(ctx context.Context, to string, report *dto.MonthlyReportDTO) error
	SendWeeklySummary(ctx context.Context, to, summary string) error
	SendBudgetAlert(ctx context.Context, to, message string) error
	SendWelcome(ctx context.Context, to, name string) error
}

// Logger is a structured logger abstraction.
type Logger interface {
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(key string, value interface{}) Logger
}
