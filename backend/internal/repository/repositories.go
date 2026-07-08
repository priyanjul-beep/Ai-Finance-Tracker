// Package repository contains all GORM-backed repository implementations.
// Each struct implements the matching interface from internal/interfaces.
package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
)

// ─── User ─────────────────────────────────────────────────────────────────────

type UserRepo struct{ db *gorm.DB }

func NewUser(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}
func (r *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error; err != nil {
		return nil, notFound("user", err)
	}
	return &u, nil
}
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error; err != nil {
		return nil, notFound("user", err)
	}
	return &u, nil
}
func (r *UserRepo) GetByGoogleID(ctx context.Context, id string) (*domain.User, error) {
	var u domain.User
	if err := r.db.WithContext(ctx).First(&u, "google_id = ?", id).Error; err != nil {
		return nil, notFound("user", err)
	}
	return &u, nil
}
func (r *UserRepo) Update(ctx context.Context, u *domain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}
func (r *UserRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

// ─── Session ──────────────────────────────────────────────────────────────────

type SessionRepo struct{ db *gorm.DB }

func NewSession(db *gorm.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, s *domain.Session) error {
	return r.db.WithContext(ctx).Create(s).Error
}
func (r *SessionRepo) GetByAccessToken(ctx context.Context, token string) (*domain.Session, error) {
	var s domain.Session
	if err := r.db.WithContext(ctx).First(&s, "access_token = ?", token).Error; err != nil {
		return nil, notFound("session", err)
	}
	return &s, nil
}
func (r *SessionRepo) GetByRefreshToken(ctx context.Context, token string) (*domain.Session, error) {
	var s domain.Session
	if err := r.db.WithContext(ctx).First(&s, "refresh_token = ?", token).Error; err != nil {
		return nil, notFound("session", err)
	}
	return &s, nil
}
func (r *SessionRepo) Update(ctx context.Context, s *domain.Session) error {
	return r.db.WithContext(ctx).Save(s).Error
}
func (r *SessionRepo) DeleteByID(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Session{}, "id = ?", id).Error
}
func (r *SessionRepo) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Delete(&domain.Session{}, "user_id = ?", userID).Error
}

// ─── Expense ──────────────────────────────────────────────────────────────────

type ExpenseRepo struct{ db *gorm.DB }

func NewExpense(db *gorm.DB) *ExpenseRepo { return &ExpenseRepo{db: db} }

func (r *ExpenseRepo) Create(ctx context.Context, e *domain.Expense) error {
	return r.db.WithContext(ctx).Create(e).Error
}
func (r *ExpenseRepo) GetByID(ctx context.Context, id string) (*domain.Expense, error) {
	var e domain.Expense
	if err := r.db.WithContext(ctx).Preload("Tags").First(&e, "id = ?", id).Error; err != nil {
		return nil, notFound("expense", err)
	}
	return &e, nil
}
func (r *ExpenseRepo) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Expense, int64, error) {
	var (
		rows  []domain.Expense
		total int64
	)
	base := r.db.WithContext(ctx).Model(&domain.Expense{}).Where("user_id = ?", userID)
	base.Count(&total)
	err := base.Preload("Tags").Order("date DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}
func (r *ExpenseRepo) Update(ctx context.Context, e *domain.Expense) error {
	return r.db.WithContext(ctx).Save(e).Error
}
func (r *ExpenseRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Expense{}, "id = ?", id).Error
}
func (r *ExpenseRepo) GetByDateRange(ctx context.Context, userID string, from, to time.Time) ([]domain.Expense, error) {
	var rows []domain.Expense
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, from, to).
		Preload("Tags").Order("date DESC").Find(&rows).Error
	return rows, err
}
func (r *ExpenseRepo) GetByCategory(ctx context.Context, userID, category string) ([]domain.Expense, error) {
	var rows []domain.Expense
	err := r.db.WithContext(ctx).Where("user_id = ? AND category = ?", userID, category).Find(&rows).Error
	return rows, err
}
func (r *ExpenseRepo) GetByMerchant(ctx context.Context, userID, merchant string) ([]domain.Expense, error) {
	var rows []domain.Expense
	err := r.db.WithContext(ctx).Where("user_id = ? AND merchant ILIKE ?", userID, "%"+merchant+"%").Find(&rows).Error
	return rows, err
}
func (r *ExpenseRepo) FindDuplicates(ctx context.Context, userID string, amount float64, merchant string) ([]domain.Expense, error) {
	var rows []domain.Expense
	// Match same merchant, ±5% amount, within 7 days
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND merchant = ? AND amount BETWEEN ? AND ? AND date >= NOW() - INTERVAL '7 days'",
			userID, merchant, amount*0.95, amount*1.05).
		Find(&rows).Error
	return rows, err
}
func (r *ExpenseRepo) SumByCategory(ctx context.Context, userID string, from, to time.Time) ([]dto.CategorySpend, error) {
	var rows []dto.CategorySpend
	err := r.db.WithContext(ctx).
		Model(&domain.Expense{}).
		Select("category, SUM(amount) as amount").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, from, to).
		Group("category").
		Scan(&rows).Error
	return rows, err
}
func (r *ExpenseRepo) SumByMerchant(ctx context.Context, userID string, from, to time.Time) ([]dto.MerchantSpend, error) {
	var rows []dto.MerchantSpend
	err := r.db.WithContext(ctx).
		Model(&domain.Expense{}).
		Select("merchant, SUM(amount) as amount").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, from, to).
		Group("merchant").Order("amount DESC").Limit(10).
		Scan(&rows).Error
	return rows, err
}
func (r *ExpenseRepo) TotalByDateRange(ctx context.Context, userID string, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&domain.Expense{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND date BETWEEN ? AND ? AND expense_type = 'spend'", userID, from, to).
		Scan(&total).Error
	return total, err
}

// ─── Income ───────────────────────────────────────────────────────────────────

type IncomeRepo struct{ db *gorm.DB }

func NewIncome(db *gorm.DB) *IncomeRepo { return &IncomeRepo{db: db} }

func (r *IncomeRepo) Create(ctx context.Context, inc *domain.Income) error {
	return r.db.WithContext(ctx).Create(inc).Error
}
func (r *IncomeRepo) GetByID(ctx context.Context, id string) (*domain.Income, error) {
	var inc domain.Income
	if err := r.db.WithContext(ctx).First(&inc, "id = ?", id).Error; err != nil {
		return nil, notFound("income", err)
	}
	return &inc, nil
}
func (r *IncomeRepo) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Income, int64, error) {
	var (
		rows  []domain.Income
		total int64
	)
	base := r.db.WithContext(ctx).Model(&domain.Income{}).Where("user_id = ?", userID)
	base.Count(&total)
	err := base.Order("date DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}
func (r *IncomeRepo) Update(ctx context.Context, inc *domain.Income) error {
	return r.db.WithContext(ctx).Save(inc).Error
}
func (r *IncomeRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Income{}, "id = ?", id).Error
}
func (r *IncomeRepo) GetByDateRange(ctx context.Context, userID string, from, to time.Time) ([]domain.Income, error) {
	var rows []domain.Income
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, from, to).
		Order("date DESC").Find(&rows).Error
	return rows, err
}
func (r *IncomeRepo) TotalByDateRange(ctx context.Context, userID string, from, to time.Time) (float64, error) {
	var total float64
	err := r.db.WithContext(ctx).
		Model(&domain.Income{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, from, to).
		Scan(&total).Error
	return total, err
}

// ─── Budget ───────────────────────────────────────────────────────────────────

type BudgetRepo struct{ db *gorm.DB }

func NewBudget(db *gorm.DB) *BudgetRepo { return &BudgetRepo{db: db} }

func (r *BudgetRepo) Create(ctx context.Context, b *domain.Budget) error {
	return r.db.WithContext(ctx).Create(b).Error
}
func (r *BudgetRepo) GetByID(ctx context.Context, id string) (*domain.Budget, error) {
	var b domain.Budget
	if err := r.db.WithContext(ctx).First(&b, "id = ?", id).Error; err != nil {
		return nil, notFound("budget", err)
	}
	return &b, nil
}
func (r *BudgetRepo) GetByUserID(ctx context.Context, userID string) ([]domain.Budget, error) {
	var rows []domain.Budget
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = true", userID).Find(&rows).Error
	return rows, err
}
func (r *BudgetRepo) GetByUserAndCategory(ctx context.Context, userID, category string) (*domain.Budget, error) {
	var b domain.Budget
	err := r.db.WithContext(ctx).First(&b, "user_id = ? AND category = ? AND is_active = true", userID, category).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &b, err
}
func (r *BudgetRepo) Update(ctx context.Context, b *domain.Budget) error {
	return r.db.WithContext(ctx).Save(b).Error
}
func (r *BudgetRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Budget{}, "id = ?", id).Error
}

// ─── Subscription ─────────────────────────────────────────────────────────────

type SubscriptionRepo struct{ db *gorm.DB }

func NewSubscription(db *gorm.DB) *SubscriptionRepo { return &SubscriptionRepo{db: db} }

func (r *SubscriptionRepo) Create(ctx context.Context, s *domain.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}
func (r *SubscriptionRepo) GetByID(ctx context.Context, id string) (*domain.Subscription, error) {
	var s domain.Subscription
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, notFound("subscription", err)
	}
	return &s, nil
}
func (r *SubscriptionRepo) GetByUserID(ctx context.Context, userID string) ([]domain.Subscription, error) {
	var rows []domain.Subscription
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = true", userID).
		Order("next_billing_date ASC").Find(&rows).Error
	return rows, err
}
func (r *SubscriptionRepo) Update(ctx context.Context, s *domain.Subscription) error {
	return r.db.WithContext(ctx).Save(s).Error
}
func (r *SubscriptionRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Subscription{}, "id = ?", id).Error
}
func (r *SubscriptionRepo) GetDueSubscriptions(ctx context.Context) ([]domain.Subscription, error) {
	var rows []domain.Subscription
	err := r.db.WithContext(ctx).
		Where("is_active = true AND next_billing_date <= NOW()").Find(&rows).Error
	return rows, err
}

// ─── Goal ─────────────────────────────────────────────────────────────────────

type GoalRepo struct{ db *gorm.DB }

func NewGoal(db *gorm.DB) *GoalRepo { return &GoalRepo{db: db} }

func (r *GoalRepo) Create(ctx context.Context, g *domain.Goal) error {
	return r.db.WithContext(ctx).Create(g).Error
}
func (r *GoalRepo) GetByID(ctx context.Context, id string) (*domain.Goal, error) {
	var g domain.Goal
	if err := r.db.WithContext(ctx).First(&g, "id = ?", id).Error; err != nil {
		return nil, notFound("goal", err)
	}
	return &g, nil
}
func (r *GoalRepo) GetByUserID(ctx context.Context, userID string) ([]domain.Goal, error) {
	var rows []domain.Goal
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("priority DESC").Find(&rows).Error
	return rows, err
}
func (r *GoalRepo) Update(ctx context.Context, g *domain.Goal) error {
	return r.db.WithContext(ctx).Save(g).Error
}
func (r *GoalRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Goal{}, "id = ?", id).Error
}

// ─── Tag ──────────────────────────────────────────────────────────────────────

type TagRepo struct{ db *gorm.DB }

func NewTag(db *gorm.DB) *TagRepo { return &TagRepo{db: db} }

func (r *TagRepo) Create(ctx context.Context, t *domain.Tag) error {
	return r.db.WithContext(ctx).Create(t).Error
}
func (r *TagRepo) GetByID(ctx context.Context, id string) (*domain.Tag, error) {
	var t domain.Tag
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, notFound("tag", err)
	}
	return &t, nil
}
func (r *TagRepo) GetByUserID(ctx context.Context, userID string) ([]domain.Tag, error) {
	var rows []domain.Tag
	return rows, r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error
}
func (r *TagRepo) Update(ctx context.Context, t *domain.Tag) error {
	return r.db.WithContext(ctx).Save(t).Error
}
func (r *TagRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Tag{}, "id = ?", id).Error
}
func (r *TagRepo) AddToExpense(ctx context.Context, tagID, expenseID string) error {
	return r.db.WithContext(ctx).
		Exec("INSERT INTO expense_tags (expense_id, tag_id) VALUES (?, ?) ON CONFLICT DO NOTHING", expenseID, tagID).Error
}
func (r *TagRepo) RemoveFromExpense(ctx context.Context, tagID, expenseID string) error {
	return r.db.WithContext(ctx).
		Exec("DELETE FROM expense_tags WHERE expense_id = ? AND tag_id = ?", expenseID, tagID).Error
}

// ─── Notification ─────────────────────────────────────────────────────────────

type NotificationRepo struct{ db *gorm.DB }

func NewNotification(db *gorm.DB) *NotificationRepo { return &NotificationRepo{db: db} }

func (r *NotificationRepo) Create(ctx context.Context, n *domain.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}
func (r *NotificationRepo) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Notification, int64, error) {
	var (
		rows  []domain.Notification
		total int64
	)
	base := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID)
	base.Count(&total)
	err := base.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}
func (r *NotificationRepo) MarkAsRead(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ?", id).Update("is_read", true).Error
}
func (r *NotificationRepo) MarkAllAsRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&domain.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error
}
func (r *NotificationRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.Notification{}, "id = ?", id).Error
}

// ─── AuditLog ─────────────────────────────────────────────────────────────────

type AuditLogRepo struct{ db *gorm.DB }

func NewAuditLog(db *gorm.DB) *AuditLogRepo { return &AuditLogRepo{db: db} }

func (r *AuditLogRepo) Create(ctx context.Context, l *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(l).Error
}
func (r *AuditLogRepo) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.AuditLog, int64, error) {
	var (
		rows  []domain.AuditLog
		total int64
	)
	base := r.db.WithContext(ctx).Model(&domain.AuditLog{}).Where("user_id = ?", userID)
	base.Count(&total)
	err := base.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error
	return rows, total, err
}

// ─── MerchantMapping ──────────────────────────────────────────────────────────

type MerchantMappingRepo struct{ db *gorm.DB }

func NewMerchantMapping(db *gorm.DB) *MerchantMappingRepo { return &MerchantMappingRepo{db: db} }

func (r *MerchantMappingRepo) Create(ctx context.Context, m *domain.MerchantMapping) error {
	return r.db.WithContext(ctx).Create(m).Error
}
func (r *MerchantMappingRepo) GetByMerchant(ctx context.Context, merchant string) (*domain.MerchantMapping, error) {
	var m domain.MerchantMapping
	err := r.db.WithContext(ctx).First(&m, "merchant ILIKE ?", merchant).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &m, err
}
func (r *MerchantMappingRepo) Update(ctx context.Context, m *domain.MerchantMapping) error {
	return r.db.WithContext(ctx).Save(m).Error
}
func (r *MerchantMappingRepo) List(ctx context.Context) ([]domain.MerchantMapping, error) {
	var rows []domain.MerchantMapping
	return rows, r.db.WithContext(ctx).Find(&rows).Error
}

// ─── RecurringExpense ─────────────────────────────────────────────────────────

type RecurringExpenseRepo struct{ db *gorm.DB }

func NewRecurringExpense(db *gorm.DB) *RecurringExpenseRepo { return &RecurringExpenseRepo{db: db} }

func (r *RecurringExpenseRepo) Create(ctx context.Context, e *domain.RecurringExpense) error {
	return r.db.WithContext(ctx).Create(e).Error
}
func (r *RecurringExpenseRepo) GetByID(ctx context.Context, id string) (*domain.RecurringExpense, error) {
	var e domain.RecurringExpense
	if err := r.db.WithContext(ctx).First(&e, "id = ?", id).Error; err != nil {
		return nil, notFound("recurring_expense", err)
	}
	return &e, nil
}
func (r *RecurringExpenseRepo) GetByUserID(ctx context.Context, userID string) ([]domain.RecurringExpense, error) {
	var rows []domain.RecurringExpense
	return rows, r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error
}
func (r *RecurringExpenseRepo) Update(ctx context.Context, e *domain.RecurringExpense) error {
	return r.db.WithContext(ctx).Save(e).Error
}
func (r *RecurringExpenseRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&domain.RecurringExpense{}, "id = ?", id).Error
}

// ─── FinancialHealthScore ─────────────────────────────────────────────────────

type HealthScoreRepo struct{ db *gorm.DB }

func NewHealthScore(db *gorm.DB) *HealthScoreRepo { return &HealthScoreRepo{db: db} }

func (r *HealthScoreRepo) Upsert(ctx context.Context, s *domain.FinancialHealthScore) error {
	return r.db.WithContext(ctx).Save(s).Error
}
func (r *HealthScoreRepo) GetByUserID(ctx context.Context, userID string) (*domain.FinancialHealthScore, error) {
	var s domain.FinancialHealthScore
	err := r.db.WithContext(ctx).First(&s, "user_id = ?", userID).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &s, err
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func notFound(entity string, err error) error {
	if err == gorm.ErrRecordNotFound {
		return fmt.Errorf("%s not found", entity)
	}
	return err
}
