// Package dto contains all Data Transfer Objects used across the API boundary.
// Every request body, response body, and inter-layer transfer goes through a DTO.
package dto

import (
	"time"
)

// ─── Auth ─────────────────────────────────────────────────────────────────────

type SignupRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type PasswordResetConfirm struct {
	Token    string `json:"token"    validate:"required"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type AuthResponse struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresIn    int64   `json:"expires_in"` // seconds
	TokenType    string  `json:"token_type"` // Bearer
	User         UserDTO `json:"user"`
}

// ─── User ─────────────────────────────────────────────────────────────────────

type UserDTO struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	ProfilePicture    string    `json:"profile_picture,omitempty"`
	IsEmailVerified   bool      `json:"is_email_verified"`
	Timezone          string    `json:"timezone"`
	Currency          string    `json:"currency"`
	PreferredLanguage string    `json:"preferred_language"`
	CreatedAt         time.Time `json:"created_at"`
}

type UpdateUserRequest struct {
	Name              string `json:"name"               validate:"omitempty,min=2,max=100"`
	Timezone          string `json:"timezone"           validate:"omitempty"`
	Currency          string `json:"currency"           validate:"omitempty"`
	PreferredLanguage string `json:"preferred_language" validate:"omitempty"`
	ProfilePicture    string `json:"profile_picture"    validate:"omitempty,url"`
}

// ─── Expense ──────────────────────────────────────────────────────────────────

type CreateExpenseRequest struct {
	Amount        float64   `json:"amount"         validate:"required,gt=0"`
	Category      string    `json:"category"       validate:"required"`
	Merchant      string    `json:"merchant"       validate:"required"`
	Description   string    `json:"description"    validate:"omitempty,max=500"`
	Notes         string    `json:"notes"          validate:"omitempty,max=1000"`
	Date          time.Time `json:"date"           validate:"required"`
	ExpenseType   string    `json:"expense_type"   validate:"omitempty,oneof=spend refund transfer"`
	PaymentMethod string    `json:"payment_method" validate:"omitempty,oneof=cash card upi bank wallet online"`
	Tags          []string  `json:"tags"           validate:"omitempty"`
}

type UpdateExpenseRequest struct {
	Amount        float64   `json:"amount"         validate:"omitempty,gt=0"`
	Category      string    `json:"category"       validate:"omitempty"`
	Merchant      string    `json:"merchant"       validate:"omitempty"`
	Description   string    `json:"description"    validate:"omitempty,max=500"`
	Notes         string    `json:"notes"          validate:"omitempty,max=1000"`
	Date          time.Time `json:"date"           validate:"omitempty"`
	PaymentMethod string    `json:"payment_method" validate:"omitempty,oneof=cash card upi bank wallet online"`
	Tags          []string  `json:"tags"           validate:"omitempty"`
	IsFavorite    *bool     `json:"is_favorite"    validate:"omitempty"`
}

type ExpenseDTO struct {
	ID            string    `json:"id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Category      string    `json:"category"`
	Merchant      string    `json:"merchant"`
	Description   string    `json:"description"`
	Notes         string    `json:"notes"`
	Date          time.Time `json:"date"`
	ExpenseType   string    `json:"expense_type"`
	PaymentMethod string    `json:"payment_method"`
	ImageURL      string    `json:"image_url,omitempty"`
	IsFavorite    bool      `json:"is_favorite"`
	Tags          []TagDTO  `json:"tags,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AIExpenseParseRequest struct {
	Text     string `json:"text"      validate:"required_without=ImageURL"`
	ImageURL string `json:"image_url" validate:"omitempty,url"`
}

type AIExpenseParseResponse struct {
	Amount        float64   `json:"amount"`
	Merchant      string    `json:"merchant"`
	Category      string    `json:"category"`
	Date          time.Time `json:"date"`
	Notes         string    `json:"notes"`
	ExpenseType   string    `json:"expense_type"`
	PaymentMethod string    `json:"payment_method"`
	Confidence    float64   `json:"confidence"` // 0-1
}

// AIVoiceParseResponse is returned by the voice-parse endpoint.
// DateStr is a raw string ("today", "yesterday", "2026-07-10") so the
// frontend can resolve it relative to the user's local clock.
type AIVoiceParseResponse struct {
	Transcript    string  `json:"transcript"`
	Amount        float64 `json:"amount"`
	Merchant      string  `json:"merchant"`
	Category      string  `json:"category"`
	DateStr       string  `json:"date"`
	Notes         string  `json:"notes"`
	ExpenseType   string  `json:"expense_type"`
	PaymentMethod string  `json:"payment_method"`
	Confidence    float64 `json:"confidence"`
	Cached        bool    `json:"cached"`
}

// ─── Income ───────────────────────────────────────────────────────────────────

type CreateIncomeRequest struct {
	Amount        float64   `json:"amount"         validate:"required,gt=0"`
	Source        string    `json:"source"         validate:"required"`
	Category      string    `json:"category"       validate:"omitempty"`
	Description   string    `json:"description"    validate:"omitempty,max=500"`
	Notes         string    `json:"notes"          validate:"omitempty,max=1000"`
	Date          time.Time `json:"date"           validate:"required"`
	PaymentMethod string    `json:"payment_method" validate:"omitempty"`
	IsTaxable     bool      `json:"is_taxable"`
	TaxAmount     float64   `json:"tax_amount"     validate:"omitempty,gte=0"`
}

type IncomeDTO struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Source        string    `json:"source"`
	Category      string    `json:"category"`
	Description   string    `json:"description"`
	Notes         string    `json:"notes"`
	Date          time.Time `json:"date"`
	PaymentMethod string    `json:"payment_method"`
	IsTaxable     bool      `json:"is_taxable"`
	TaxAmount     float64   `json:"tax_amount"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpdateIncomeRequest for partial income updates.
type UpdateIncomeRequest struct {
	Amount        float64   `json:"amount"         validate:"omitempty,gt=0"`
	Source        string    `json:"source"         validate:"omitempty"`
	Category      string    `json:"category"       validate:"omitempty"`
	Description   string    `json:"description"    validate:"omitempty,max=500"`
	Notes         string    `json:"notes"          validate:"omitempty,max=1000"`
	Date          time.Time `json:"date"`
	PaymentMethod string    `json:"payment_method" validate:"omitempty"`
	IsTaxable     bool      `json:"is_taxable"`
	TaxAmount     float64   `json:"tax_amount"     validate:"omitempty,gte=0"`
}

// ─── Budget ───────────────────────────────────────────────────────────────────

type CreateBudgetRequest struct {
	Category    string  `json:"category"    validate:"required"`
	Amount      float64 `json:"amount"      validate:"required,gt=0"`
	Period      string  `json:"period"      validate:"required,oneof=monthly yearly weekly"`
	Month       int     `json:"month"       validate:"omitempty,min=0,max=12"`
	Year        int     `json:"year"        validate:"required,min=2000,max=2100"`
	AlertAt     float64 `json:"alert_at"    validate:"omitempty,min=1,max=100"`
	Description string  `json:"description" validate:"omitempty,max=500"`
}

type BudgetDTO struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Category    string    `json:"category"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Period      string    `json:"period"`
	Month       int       `json:"month"`
	Year        int       `json:"year"`
	AlertAt     float64   `json:"alert_at"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BudgetStatusDTO struct {
	BudgetDTO
	Spent     float64 `json:"spent"`
	Remaining float64 `json:"remaining"`
	Percent   float64 `json:"percent"`
	Status    string  `json:"status"` // on-track | warning | over-budget
}

// UpdateBudgetRequest for partial budget updates.
type UpdateBudgetRequest struct {
	Amount      float64 `json:"amount"      validate:"omitempty,gt=0"`
	AlertAt     float64 `json:"alert_at"    validate:"omitempty,min=1,max=100"`
	Description string  `json:"description" validate:"omitempty,max=500"`
	Period      string  `json:"period"      validate:"omitempty"`
}

// ─── Subscription ─────────────────────────────────────────────────────────────

type CreateSubscriptionRequest struct {
	Name            string    `json:"name"              validate:"required"`
	Amount          float64   `json:"amount"            validate:"required,gt=0"`
	BillingCycle    string    `json:"billing_cycle"     validate:"required,oneof=daily weekly monthly yearly"`
	NextBillingDate time.Time `json:"next_billing_date" validate:"required"`
	Category        string    `json:"category"          validate:"omitempty"`
	PaymentMethod   string    `json:"payment_method"    validate:"omitempty"`
	Notes           string    `json:"notes"             validate:"omitempty,max=500"`
}

type SubscriptionDTO struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Name            string    `json:"name"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	BillingCycle    string    `json:"billing_cycle"`
	NextBillingDate time.Time `json:"next_billing_date"`
	Category        string    `json:"category"`
	PaymentMethod   string    `json:"payment_method"`
	Notes           string    `json:"notes"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UpdateSubscriptionRequest for partial subscription updates.
type UpdateSubscriptionRequest struct {
	Name            string    `json:"name"            validate:"omitempty,min=2,max=100"`
	Amount          float64   `json:"amount"          validate:"omitempty,gt=0"`
	BillingCycle    string    `json:"billing_cycle"   validate:"omitempty,oneof=daily weekly monthly yearly"`
	NextBillingDate time.Time `json:"next_billing_date"`
	Category        string    `json:"category"        validate:"omitempty"`
	PaymentMethod   string    `json:"payment_method"  validate:"omitempty"`
	Notes           string    `json:"notes"           validate:"omitempty,max=500"`
	IsActive        *bool     `json:"is_active"`
}

// ─── Goal ─────────────────────────────────────────────────────────────────────

type CreateGoalRequest struct {
	Name         string    `json:"name"          validate:"required,min=2,max=100"`
	Description  string    `json:"description"   validate:"omitempty,max=500"`
	TargetAmount float64   `json:"target_amount" validate:"required,gt=0"`
	Category     string    `json:"category"      validate:"omitempty"`
	TargetDate   time.Time `json:"target_date"   validate:"required"`
	Priority     int       `json:"priority"      validate:"omitempty,min=1,max=5"`
}

type GoalDTO struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	TargetAmount         float64   `json:"target_amount"`
	CurrentAmount        float64   `json:"current_amount"`
	Currency             string    `json:"currency"`
	Category             string    `json:"category"`
	TargetDate           time.Time `json:"target_date"`
	Priority             int       `json:"priority"`
	Status               string    `json:"status"`
	ProgressPercent      float64   `json:"progress_percent"` // 0-100
	DaysRemaining        int       `json:"days_remaining"`
	MonthlySavingsNeeded float64   `json:"monthly_savings_needed"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// UpdateGoalRequest for partial goal updates.
type UpdateGoalRequest struct {
	Name          string    `json:"name"          validate:"omitempty,min=2,max=100"`
	Description   string    `json:"description"   validate:"omitempty,max=500"`
	TargetAmount  float64   `json:"target_amount" validate:"omitempty,gt=0"`
	TargetDate    time.Time `json:"target_date"`
	Priority      int       `json:"priority"      validate:"omitempty,min=1,max=5"`
	Status        string    `json:"status"        validate:"omitempty,oneof=active completed abandoned paused"`
}

// ─── Tag ──────────────────────────────────────────────────────────────────────

type CreateTagRequest struct {
	Name  string `json:"name"  validate:"required,min=1,max=50"`
	Color string `json:"color" validate:"required,hexcolor"`
}

type TagDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// UpdateTagRequest for partial tag updates.
type UpdateTagRequest struct {
	Name  string `json:"name"  validate:"omitempty,min=1,max=50"`
	Color string `json:"color" validate:"omitempty,hexcolor"`
}

// ─── Notification ─────────────────────────────────────────────────────────────

type NotificationDTO struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// ─── Analytics / Dashboard ────────────────────────────────────────────────────

type DashboardDTO struct {
	CurrentBalance       float64            `json:"total_balance"`
	TotalIncome          float64            `json:"total_income"`
	TotalExpense         float64            `json:"total_expenses"`
	TotalSavings         float64            `json:"total_savings"`
	SavingsRate          float64            `json:"savings_rate"`
	MonthlySpending      float64            `json:"monthly_spend"`
	WeeklySpending       float64            `json:"weekly_spend"`
	FinancialHealthScore float64            `json:"financial_health_score"`
	CategoryBreakdown    []CategorySpend    `json:"spend_by_category"`
	RecentExpenses       []ExpenseDTO       `json:"recent_expenses"`
	UpcomingBills        []SubscriptionDTO  `json:"upcoming_subscriptions"`
	TopMerchants         []MerchantSpend    `json:"spend_by_merchant"`
	Predictions          PredictionData     `json:"predictions"`
}

type CategorySpend struct {
	Category   string  `json:"category"`
	Amount     float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type MerchantSpend struct {
	Merchant   string  `json:"merchant"`
	Amount     float64 `json:"amount"`
	Category   string  `json:"category"`
	Percentage float64 `json:"percentage"`
}

type PredictionData struct {
	EndOfMonthSpending  float64 `json:"end_of_month_spending"`
	NextMonthPrediction float64 `json:"next_month_prediction"` // alias shown in dashboard
	Confidence          float64 `json:"confidence"`            // 0.0-1.0
	Trend               string  `json:"trend"`                 // increasing|decreasing|stable
	ExpectedSavings     float64 `json:"expected_savings"`
	BudgetOverrunRisk   float64 `json:"budget_overrun_risk"` // 0-100
	SavingsGoalOnTrack  bool    `json:"savings_goal_on_track"`
}

type MonthlyReportDTO struct {
	Month             int             `json:"month"`
	Year              int             `json:"year"`
	TotalIncome       float64         `json:"total_income"`
	TotalExpense      float64         `json:"total_expense"`
	TotalSavings      float64         `json:"total_savings"`
	SavingsRate       float64         `json:"savings_rate"` // %
	TopCategories     []CategorySpend `json:"top_categories"`
	TopMerchants      []MerchantSpend `json:"top_merchants"`
	BudgetHealth      float64         `json:"budget_health"`
	FinancialScore    float64         `json:"financial_score"`
	Recommendations   []string        `json:"recommendations"`
}

// ─── Pagination ───────────────────────────────────────────────────────────────

type PaginationParams struct {
	Page  int    `form:"page"  validate:"required,min=1"`
	Limit int    `form:"limit" validate:"required,min=1,max=100"`
	Sort  string `form:"sort"`
	Order string `form:"order" validate:"omitempty,oneof=asc desc"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// ─── Error ────────────────────────────────────────────────────────────────────

type ErrorResponse struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ─── Report ───────────────────────────────────────────────────────────────────

type ReportRequest struct {
	Type      string    `json:"type"       validate:"required,oneof=monthly yearly expense tax"`
	StartDate time.Time `json:"start_date" validate:"required"`
	EndDate   time.Time `json:"end_date"   validate:"required,gtfield=StartDate"`
	Format    string    `json:"format"     validate:"required,oneof=pdf csv"`
}
