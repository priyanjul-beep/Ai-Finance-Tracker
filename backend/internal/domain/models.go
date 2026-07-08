package domain

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ─── User ─────────────────────────────────────────────────────────────────────

// User is the central entity representing an authenticated account.
type User struct {
	ID                string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Email             string         `gorm:"uniqueIndex;not null"                           json:"email"`
	Name              string         `gorm:"not null"                                       json:"name"`
	PasswordHash      string         `gorm:"column:password_hash"                           json:"-"`
	ProfilePicture    string         `json:"profile_picture,omitempty"`
	IsEmailVerified   bool           `gorm:"default:false"                                  json:"is_email_verified"`
	EmailVerifyToken  string         `json:"-"`
	ResetToken        string         `json:"-"`
	ResetTokenExpiry  *time.Time     `json:"-"`
	GoogleID          string         `gorm:"index"                                          json:"-"`
	Timezone          string         `gorm:"default:'Asia/Kolkata'"                         json:"timezone"`
	Currency          string         `gorm:"default:'INR'"                                  json:"currency"`
	PreferredLanguage string         `gorm:"default:'en'"                                   json:"preferred_language"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index"                                          json:"-"`
}

func (User) TableName() string { return "users" }

// ─── Enums ────────────────────────────────────────────────────────────────────

type ExpenseType    string
type PaymentMethod  string
type ExpenseCategory string

const (
	ExpenseTypeSpend    ExpenseType = "spend"
	ExpenseTypeRefund   ExpenseType = "refund"
	ExpenseTypeTransfer ExpenseType = "transfer"

	PaymentCash   PaymentMethod = "cash"
	PaymentCard   PaymentMethod = "card"
	PaymentUPI    PaymentMethod = "upi"
	PaymentBank   PaymentMethod = "bank"
	PaymentWallet PaymentMethod = "wallet"
	PaymentOnline PaymentMethod = "online"

	CatFood          ExpenseCategory = "food"
	CatTravel        ExpenseCategory = "travel"
	CatShopping      ExpenseCategory = "shopping"
	CatEntertainment ExpenseCategory = "entertainment"
	CatHealth        ExpenseCategory = "health"
	CatInvestment    ExpenseCategory = "investment"
	CatEducation     ExpenseCategory = "education"
	CatBills         ExpenseCategory = "bills"
	CatRecharge      ExpenseCategory = "recharge"
	CatFuel          ExpenseCategory = "fuel"
	CatRent          ExpenseCategory = "rent"
	CatSalary        ExpenseCategory = "salary"
	CatUtilities     ExpenseCategory = "utilities"
	CatSubscription  ExpenseCategory = "subscription"
	CatPersonalCare  ExpenseCategory = "personal_care"
	CatGift          ExpenseCategory = "gift"
	CatCharitable    ExpenseCategory = "charitable"
	CatInsurance     ExpenseCategory = "insurance"
	CatOthers        ExpenseCategory = "others"
	CatUnknown       ExpenseCategory = "unknown"
)

// ─── Expense ──────────────────────────────────────────────────────────────────

// Expense records a single spend/refund/transfer event.
type Expense struct {
	ID            string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID        string          `gorm:"index;not null"                                 json:"user_id"`
	Amount        float64         `gorm:"not null"                                       json:"amount"`
	Currency      string          `gorm:"default:'INR'"                                  json:"currency"`
	Category      ExpenseCategory `gorm:"not null"                                       json:"category"`
	Merchant      string          `json:"merchant"`
	Description   string          `json:"description"`
	Notes         string          `json:"notes"`
	Date          time.Time       `gorm:"index"                                          json:"date"`
	ExpenseType   ExpenseType     `gorm:"default:'spend'"                                json:"expense_type"`
	PaymentMethod PaymentMethod   `json:"payment_method"`
	ImageURL      string          `json:"image_url,omitempty"`
	OCRData       datatypes.JSON  `json:"ocr_data,omitempty"`
	IsDuplicate   bool            `gorm:"default:false"                                  json:"is_duplicate"`
	DuplicateOf   *string         `json:"duplicate_of,omitempty"`
	IsFavorite    bool            `gorm:"default:false"                                  json:"is_favorite"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `gorm:"index"                                          json:"-"`

	User User  `gorm:"foreignKey:UserID" json:"-"`
	Tags []Tag `gorm:"many2many:expense_tags;"   json:"tags,omitempty"`
}

func (Expense) TableName() string { return "expenses" }

// ─── Income ───────────────────────────────────────────────────────────────────

// Income records a single income event (salary, freelance, bonus, etc.).
type Income struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID        string         `gorm:"index;not null"                                 json:"user_id"`
	Amount        float64        `gorm:"not null"                                       json:"amount"`
	Currency      string         `gorm:"default:'INR'"                                  json:"currency"`
	Source        string         `json:"source"`
	Category      string         `json:"category"`
	Description   string         `json:"description"`
	Notes         string         `json:"notes"`
	Date          time.Time      `gorm:"index"                                          json:"date"`
	PaymentMethod PaymentMethod  `json:"payment_method"`
	IsTaxable     bool           `gorm:"default:false"                                  json:"is_taxable"`
	TaxAmount     float64        `json:"tax_amount"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Income) TableName() string { return "incomes" }

// ─── Budget ───────────────────────────────────────────────────────────────────

// Budget defines a spending limit for a category within a period.
type Budget struct {
	ID          string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID      string          `gorm:"index;not null"                                 json:"user_id"`
	Category    ExpenseCategory `gorm:"not null"                                       json:"category"`
	Amount      float64         `gorm:"not null"                                       json:"amount"`
	Currency    string          `gorm:"default:'INR'"                                  json:"currency"`
	Period      string          `gorm:"default:'monthly'"                              json:"period"` // monthly | yearly | weekly
	Month       int             `json:"month"`                                                        // 1-12; 0 for yearly
	Year        int             `gorm:"not null"                                       json:"year"`
	AlertAt     float64         `gorm:"default:80"                                     json:"alert_at"` // % threshold
	Description string          `json:"description"`
	IsActive    bool            `gorm:"default:true"                                   json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Budget) TableName() string { return "budgets" }

// ─── Subscription ─────────────────────────────────────────────────────────────

// Subscription tracks recurring service payments (Netflix, Spotify, etc.).
type Subscription struct {
	ID              string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID          string          `gorm:"index;not null"                                 json:"user_id"`
	Name            string          `gorm:"not null"                                       json:"name"`
	Amount          float64         `gorm:"not null"                                       json:"amount"`
	Currency        string          `gorm:"default:'INR'"                                  json:"currency"`
	BillingCycle    string          `gorm:"default:'monthly'"                              json:"billing_cycle"` // monthly | yearly | weekly | daily
	NextBillingDate time.Time       `gorm:"index"                                          json:"next_billing_date"`
	Category        ExpenseCategory `json:"category"`
	PaymentMethod   PaymentMethod   `json:"payment_method"`
	Notes           string          `json:"notes"`
	IsActive        bool            `gorm:"default:true"                                   json:"is_active"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Subscription) TableName() string { return "subscriptions" }

// ─── Goal ─────────────────────────────────────────────────────────────────────

// Goal tracks a financial savings target (emergency fund, vacation, etc.).
type Goal struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID        string         `gorm:"index;not null"                                 json:"user_id"`
	Name          string         `gorm:"not null"                                       json:"name"`
	Description   string         `json:"description"`
	TargetAmount  float64        `gorm:"not null"                                       json:"target_amount"`
	CurrentAmount float64        `gorm:"default:0"                                      json:"current_amount"`
	Currency      string         `gorm:"default:'INR'"                                  json:"currency"`
	Category      string         `json:"category"`
	TargetDate    time.Time      `json:"target_date"`
	Priority      int            `gorm:"default:3"                                      json:"priority"` // 1-5
	Status        string         `gorm:"default:'active'"                               json:"status"`   // active | achieved | abandoned
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Goal) TableName() string { return "goals" }

// ─── Tag ──────────────────────────────────────────────────────────────────────

// Tag is a user-defined label that can be attached to expenses.
type Tag struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID    string         `gorm:"index;not null"                                 json:"user_id"`
	Name      string         `gorm:"not null"                                       json:"name"`
	Color     string         `gorm:"default:'#6366f1'"                              json:"color"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`

	User     User      `gorm:"foreignKey:UserID" json:"-"`
	Expenses []Expense `gorm:"many2many:expense_tags;" json:"expenses,omitempty"`
}

func (Tag) TableName() string { return "tags" }

// ─── Notification ─────────────────────────────────────────────────────────────

// Notification stores in-app notifications for the user.
type Notification struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID    string         `gorm:"index;not null"                                 json:"user_id"`
	Title     string         `gorm:"not null"                                       json:"title"`
	Message   string         `json:"message"`
	Type      string         `json:"type"` // budget_warning | weekly_summary | monthly_report | large_expense | recurring_reminder | low_balance | goal_achieved
	IsRead    bool           `gorm:"default:false"                                  json:"is_read"`
	Data      datatypes.JSON `json:"data,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Notification) TableName() string { return "notifications" }

// ─── AuditLog ─────────────────────────────────────────────────────────────────

// AuditLog records every create/update/delete action for compliance.
type AuditLog struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID    string         `gorm:"index;not null"                                 json:"user_id"`
	Action    string         `gorm:"not null"                                       json:"action"` // CREATE | UPDATE | DELETE
	Entity    string         `gorm:"not null"                                       json:"entity"`
	EntityID  string         `gorm:"index"                                          json:"entity_id"`
	OldData   datatypes.JSON `json:"old_data,omitempty"`
	NewData   datatypes.JSON `json:"new_data,omitempty"`
	IPAddress string         `json:"ip_address"`
	UserAgent string         `json:"user_agent"`
	CreatedAt time.Time      `json:"created_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// ─── RecurringExpense ─────────────────────────────────────────────────────────

// RecurringExpense is an AI/rule-detected repeating payment pattern.
type RecurringExpense struct {
	ID             string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID         string          `gorm:"index;not null"                                 json:"user_id"`
	Merchant       string          `json:"merchant"`
	Amount         float64         `json:"amount"`
	Currency       string          `gorm:"default:'INR'"                                  json:"currency"`
	Category       ExpenseCategory `json:"category"`
	Frequency      string          `json:"frequency"` // daily | weekly | monthly | yearly
	LastOccurrence time.Time       `json:"last_occurrence"`
	NextOccurrence time.Time       `gorm:"index"                                          json:"next_occurrence"`
	Confidence     float64         `json:"confidence"` // 0-1
	IsApproved     bool            `gorm:"default:false"                                  json:"is_approved"`
	IsActive       bool            `gorm:"default:true"                                   json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	DeletedAt      gorm.DeletedAt  `gorm:"index"                                          json:"-"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (RecurringExpense) TableName() string { return "recurring_expenses" }

// ─── MerchantMapping ──────────────────────────────────────────────────────────

// MerchantMapping is a global look-up used for instant categorisation.
type MerchantMapping struct {
	ID        string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Merchant  string          `gorm:"uniqueIndex;not null"                           json:"merchant"`
	Category  ExpenseCategory `gorm:"not null"                                       json:"category"`
	Aliases   pq.StringArray  `gorm:"type:text[]"                                    json:"aliases"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (MerchantMapping) TableName() string { return "merchant_mappings" }

// ─── Session ──────────────────────────────────────────────────────────────────

// Session persists a JWT pair so tokens can be revoked server-side.
type Session struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID           string    `gorm:"index;not null"                                 json:"user_id"`
	AccessToken      string    `gorm:"index"                                          json:"-"`
	RefreshToken     string    `gorm:"uniqueIndex"                                    json:"-"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
	IPAddress        string    `json:"ip_address"`
	UserAgent        string    `json:"user_agent"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (Session) TableName() string { return "sessions" }

// ─── FinancialHealthScore ─────────────────────────────────────────────────────

// FinancialHealthScore is a computed 0-100 score refreshed periodically.
type FinancialHealthScore struct {
	ID                 string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserID             string         `gorm:"uniqueIndex;not null"                           json:"user_id"`
	Score              float64        `json:"score"`
	IncomeScore        float64        `json:"income_score"`
	SavingsScore       float64        `json:"savings_score"`
	ExpenseRatio       float64        `json:"expense_ratio"`
	BudgetHealth       float64        `json:"budget_health"`
	DebtHealth         float64        `json:"debt_health"`
	SubscriptionHealth float64        `json:"subscription_health"`
	Insights           datatypes.JSON `json:"insights,omitempty"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (FinancialHealthScore) TableName() string { return "financial_health_scores" }
