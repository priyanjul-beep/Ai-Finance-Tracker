// Package usecase – analytics / dashboard business logic.
package usecase

import (
	"context"
	"time"

	"github.com/priyanjul/ai-finance-tracker/internal/domain"
	"github.com/priyanjul/ai-finance-tracker/internal/dto"
	"github.com/priyanjul/ai-finance-tracker/internal/interfaces"
)

// AnalyticsUseCase implements interfaces.AnalyticsService.
type AnalyticsUseCase struct {
	expenses    interfaces.ExpenseRepository
	incomes     interfaces.IncomeRepository
	budgets     interfaces.BudgetRepository
	healthScore interfaces.FinancialHealthScoreRepository
	subscriptions interfaces.SubscriptionRepository
	ai          interfaces.AIProvider
	cache       interfaces.CacheService
}

// NewAnalytics creates a new AnalyticsUseCase.
func NewAnalytics(
	expenses interfaces.ExpenseRepository,
	incomes interfaces.IncomeRepository,
	budgets interfaces.BudgetRepository,
	healthScore interfaces.FinancialHealthScoreRepository,
	subscriptions interfaces.SubscriptionRepository,
	ai interfaces.AIProvider,
	cache interfaces.CacheService,
) *AnalyticsUseCase {
	return &AnalyticsUseCase{
		expenses: expenses, incomes: incomes, budgets: budgets,
		healthScore: healthScore, subscriptions: subscriptions,
		ai: ai, cache: cache,
	}
}

// GetDashboard returns the full dashboard payload, using Redis cache when available.
func (uc *AnalyticsUseCase) GetDashboard(ctx context.Context, userID string) (*dto.DashboardDTO, error) {
	var cached dto.DashboardDTO
	if err := uc.cache.GetDashboard(ctx, userID, &cached); err == nil {
		return &cached, nil
	}

	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))

	totalIncome, _ := uc.incomes.TotalByDateRange(ctx, userID, monthStart, now)
	totalExpense, _ := uc.expenses.TotalByDateRange(ctx, userID, monthStart, now)
	weeklySpend, _ := uc.expenses.TotalByDateRange(ctx, userID, weekStart, now)

	categories, _ := uc.expenses.SumByCategory(ctx, userID, monthStart, now)
	merchants, _ := uc.expenses.SumByMerchant(ctx, userID, monthStart, now)

	// Calculate percentages
	enrichCategories(categories, totalExpense)
	enrichMerchants(merchants, totalExpense)

	// Recent expenses (last 10)
	recentRows, _, _ := uc.expenses.GetByUserID(ctx, userID, 10, 0)
	var recentExpenses []dto.ExpenseDTO
	for i := range recentRows {
		recentExpenses = append(recentExpenses, *mapExpense(&recentRows[i]))
	}

	// Upcoming subscriptions
	subs, _ := uc.subscriptions.GetByUserID(ctx, userID)
	var upcoming []dto.SubscriptionDTO
	for _, s := range subs {
		if s.NextBillingDate.Before(now.AddDate(0, 0, 30)) {
			upcoming = append(upcoming, mapSubscription(&s))
		}
	}

	// Financial health score
	healthScoreVal := 0.0
	if hs, _ := uc.healthScore.GetByUserID(ctx, userID); hs != nil {
		healthScoreVal = hs.Score
	}

	// Savings rate
	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = ((totalIncome - totalExpense) / totalIncome) * 100
	}

	// AI predictions (graceful degradation: never 500 on AI failure)
	predData := map[string]interface{}{
		"total_expense":  totalExpense,
		"total_income":   totalIncome,
		"days_in_month":  now.Day(),
		"days_remaining": daysRemaining(now),
	}
	predictions, _ := uc.ai.PredictExpenses(ctx, predData)
	if predictions == nil {
		predictions = &dto.PredictionData{}
	}
	// Populate frontend-expected aliases
	predictions.NextMonthPrediction = predictions.EndOfMonthSpending
	predictions.Confidence = (100 - predictions.BudgetOverrunRisk) / 100
	if predictions.Confidence <= 0 {
		predictions.Confidence = 0.75 // sensible default when AI unavailable
	}
	switch {
	case totalExpense > totalIncome*0.85:
		predictions.Trend = "increasing"
	case totalExpense < totalIncome*0.5:
		predictions.Trend = "decreasing"
	default:
		predictions.Trend = "stable"
	}

	dashboard := &dto.DashboardDTO{
		CurrentBalance:       totalIncome - totalExpense,
		TotalIncome:          totalIncome,
		TotalExpense:         totalExpense,
		TotalSavings:         totalIncome - totalExpense,
		SavingsRate:          savingsRate,
		MonthlySpending:      totalExpense,
		WeeklySpending:       weeklySpend,
		FinancialHealthScore: healthScoreVal,
		CategoryBreakdown:    categories,
		RecentExpenses:       recentExpenses,
		UpcomingBills:        upcoming,
		TopMerchants:         merchants,
		Predictions:          *predictions,
	}

	// Cache for 30 minutes
	_ = uc.cache.SetDashboard(ctx, userID, dashboard)
	return dashboard, nil
}

// GetMonthlyReport assembles a detailed monthly report with AI recommendations.
func (uc *AnalyticsUseCase) GetMonthlyReport(ctx context.Context, userID string, month, year int) (*dto.MonthlyReportDTO, error) {
	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, 0).Add(-time.Second)

	totalIncome, _ := uc.incomes.TotalByDateRange(ctx, userID, from, to)
	totalExpense, _ := uc.expenses.TotalByDateRange(ctx, userID, from, to)
	categories, _ := uc.expenses.SumByCategory(ctx, userID, from, to)
	merchants, _ := uc.expenses.SumByMerchant(ctx, userID, from, to)

	enrichCategories(categories, totalExpense)
	enrichMerchants(merchants, totalExpense)

	savingsRate := 0.0
	if totalIncome > 0 {
		savingsRate = ((totalIncome - totalExpense) / totalIncome) * 100
	}

	healthScoreVal := 0.0
	if hs, _ := uc.healthScore.GetByUserID(ctx, userID); hs != nil {
		healthScoreVal = hs.Score
	}

	// AI recommendations
	aiData := map[string]interface{}{
		"total_income":   totalIncome,
		"total_expense":  totalExpense,
		"categories":     categories,
		"savings_rate":   savingsRate,
	}
	recommendations, _ := uc.ai.GenerateInsights(ctx, aiData)

	return &dto.MonthlyReportDTO{
		Month:           month,
		Year:            year,
		TotalIncome:     totalIncome,
		TotalExpense:    totalExpense,
		TotalSavings:    totalIncome - totalExpense,
		SavingsRate:     savingsRate,
		TopCategories:   categories,
		TopMerchants:    merchants,
		FinancialScore:  healthScoreVal,
		Recommendations: recommendations,
	}, nil
}

// GetYearlyReport returns a month-by-month yearly overview.
func (uc *AnalyticsUseCase) GetYearlyReport(ctx context.Context, userID string, year int) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	monthly := make([]map[string]interface{}, 0, 12)

	for m := 1; m <= 12; m++ {
		from := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0).Add(-time.Second)

		income, _ := uc.incomes.TotalByDateRange(ctx, userID, from, to)
		expense, _ := uc.expenses.TotalByDateRange(ctx, userID, from, to)

		monthly = append(monthly, map[string]interface{}{
			"month":   m,
			"income":  income,
			"expense": expense,
			"savings": income - expense,
		})
	}
	result["year"] = year
	result["monthly"] = monthly
	return result, nil
}

// GetPredictions returns AI-driven spending forecasts.
func (uc *AnalyticsUseCase) GetPredictions(ctx context.Context, userID string) (*dto.PredictionData, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	totalExpense, _ := uc.expenses.TotalByDateRange(ctx, userID, monthStart, now)
	totalIncome, _ := uc.incomes.TotalByDateRange(ctx, userID, monthStart, now)

	data := map[string]interface{}{
		"total_expense":  totalExpense,
		"total_income":   totalIncome,
		"days_in_month":  now.Day(),
		"days_remaining": daysRemaining(now),
	}
	return uc.ai.PredictExpenses(ctx, data)
}

// GetInsights returns AI-generated personal finance tips.
func (uc *AnalyticsUseCase) GetInsights(ctx context.Context, userID string) ([]string, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	totalExpense, _ := uc.expenses.TotalByDateRange(ctx, userID, monthStart, now)
	totalIncome, _ := uc.incomes.TotalByDateRange(ctx, userID, monthStart, now)
	categories, _ := uc.expenses.SumByCategory(ctx, userID, monthStart, now)

	data := map[string]interface{}{
		"total_expense": totalExpense,
		"total_income":  totalIncome,
		"categories":    categories,
	}
	insights, err := uc.ai.GenerateInsights(ctx, data)
	if err != nil || insights == nil {
		// Graceful degradation: return empty list instead of 500
		return []string{}, nil
	}
	return insights, nil
}

// GetFinancialHealthScore retrieves (or computes) the user's health score.
func (uc *AnalyticsUseCase) GetFinancialHealthScore(ctx context.Context, userID string) (*domain.FinancialHealthScore, error) {
	return uc.healthScore.GetByUserID(ctx, userID)
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func enrichCategories(cats []dto.CategorySpend, total float64) {
	for i := range cats {
		if total > 0 {
			cats[i].Percentage = (cats[i].Amount / total) * 100
		}
	}
}

func enrichMerchants(merchants []dto.MerchantSpend, total float64) {
	for i := range merchants {
		if total > 0 {
			merchants[i].Percentage = (merchants[i].Amount / total) * 100
		}
	}
}

func daysRemaining(t time.Time) int {
	lastDay := time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location())
	return lastDay.Day() - t.Day()
}

func mapSubscription(s *domain.Subscription) dto.SubscriptionDTO {
	return dto.SubscriptionDTO{
		ID: s.ID, Name: s.Name, Amount: s.Amount, Currency: s.Currency,
		BillingCycle: s.BillingCycle, NextBillingDate: s.NextBillingDate,
		Category: string(s.Category), PaymentMethod: string(s.PaymentMethod),
		Notes: s.Notes, IsActive: s.IsActive,
		CreatedAt: s.CreatedAt, UpdatedAt: s.UpdatedAt,
	}
}
