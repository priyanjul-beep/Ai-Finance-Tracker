// Package usecase – analytics / dashboard business logic.
package usecase

import (
	"context"
	"encoding/json"
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
	result, err := uc.ai.PredictExpenses(ctx, data)
	if err != nil || result == nil {
		// Graceful degradation: return a sensible default when AI is unavailable
		trend := "stable"
		if totalExpense > totalIncome*0.85 {
			trend = "increasing"
		} else if totalExpense < totalIncome*0.5 {
			trend = "decreasing"
		}
		daysLeft := float64(daysRemaining(now))
		daysPassed := float64(now.Day())
		nextMonthEst := 0.0
		if daysPassed > 0 {
			nextMonthEst = (totalExpense / daysPassed) * (daysPassed + daysLeft)
		}
		return &dto.PredictionData{
			NextMonthPrediction: nextMonthEst,
			Confidence:          0.75,
			Trend:               trend,
		}, nil
	}
	return result, nil
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

// GetFinancialHealthScore computes (or refreshes) the user's financial health score
// from live data, persists it, and returns it.  It never returns all-zeros for
// an existing user who has financial data.
func (uc *AnalyticsUseCase) GetFinancialHealthScore(ctx context.Context, userID string) (*domain.FinancialHealthScore, error) {
	// ── 1. Gather real financial data ────────────────────────────────────────
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonthStart := monthStart.AddDate(0, -1, 0)
	prevMonthEnd := monthStart.Add(-time.Second)
	yearStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())

	totalIncome, _ := uc.incomes.TotalByDateRange(ctx, userID, monthStart, now)
	totalExpense, _ := uc.expenses.TotalByDateRange(ctx, userID, monthStart, now)
	prevIncome, _ := uc.incomes.TotalByDateRange(ctx, userID, prevMonthStart, prevMonthEnd)
	prevExpense, _ := uc.expenses.TotalByDateRange(ctx, userID, prevMonthStart, prevMonthEnd)
	yearlyIncome, _ := uc.incomes.TotalByDateRange(ctx, userID, yearStart, now)

	budgets, _ := uc.budgets.GetByUserID(ctx, userID)
	subs, _ := uc.subscriptions.GetByUserID(ctx, userID)

	// ── 2. Compute sub-scores (all 0–100) ───────────────────────────────────

	// Income Score: based on income consistency month-over-month
	incomeScore := calcIncomeScore(totalIncome, prevIncome, yearlyIncome)

	// Savings Score: savings rate this month
	savingsScore := calcSavingsScore(totalIncome, totalExpense)

	// Expense Ratio: how much of income is spent (lower = healthier)
	expenseRatio := calcExpenseRatio(totalIncome, totalExpense)

	// Budget Health: % of active budgets that are on-track
	budgetHealth := calcBudgetHealth(ctx, uc, userID, budgets, monthStart, now)

	// Subscription Health: subscription spend vs total income
	subscriptionHealth := calcSubscriptionHealth(totalIncome, subs)

	// Debt Health: placeholder — no debt module yet; use expense vs income trend
	debtHealth := calcDebtHealth(totalExpense, prevExpense, totalIncome)

	// Overall score: weighted average
	overallScore := (incomeScore*0.25 +
		savingsScore*0.30 +
		(100-expenseRatio)*0.20 +
		budgetHealth*0.15 +
		subscriptionHealth*0.05 +
		debtHealth*0.05)

	// Clamp to [0, 100]
	overallScore = clamp(overallScore, 0, 100)

	// ── 3. Optionally refine with AI (non-blocking, best-effort) ────────────
	aiData := map[string]interface{}{
		"total_income":        totalIncome,
		"total_expense":       totalExpense,
		"prev_income":         prevIncome,
		"prev_expense":        prevExpense,
		"savings_rate":        savingsRate(totalIncome, totalExpense),
		"budget_count":        len(budgets),
		"subscription_count":  len(subs),
	}
	if aiScore, err := uc.ai.CalcHealthScore(ctx, aiData); err == nil && aiScore > 0 {
		// Blend AI score (40%) with our deterministic score (60%)
		overallScore = clamp(overallScore*0.6+aiScore*0.4, 0, 100)
	}

	// ── 4. Build insights ───────────────────────────────────────────────────
	insights := buildInsights(totalIncome, totalExpense, savingsScore, budgetHealth, subscriptionHealth)

	// ── 5. Persist (upsert) ─────────────────────────────────────────────────
	hs := &domain.FinancialHealthScore{
		UserID:             userID,
		Score:              round2(overallScore),
		IncomeScore:        round2(incomeScore),
		SavingsScore:       round2(savingsScore),
		ExpenseRatio:       round2(expenseRatio),
		BudgetHealth:       round2(budgetHealth),
		DebtHealth:         round2(debtHealth),
		SubscriptionHealth: round2(subscriptionHealth),
	}
	if len(insights) > 0 {
		if b, err := marshalInsights(insights); err == nil {
			hs.Insights = b
		}
	}

	// Upsert — if record not found yet we need a fresh ID; GORM Save handles it.
	existing, _ := uc.healthScore.GetByUserID(ctx, userID)
	if existing != nil {
		hs.ID = existing.ID // preserve PK for update
	}
	_ = uc.healthScore.Upsert(ctx, hs)

	return hs, nil
}

// ─── score helper functions ────────────────────────────────────────────────────

func calcIncomeScore(income, prevIncome, yearlyIncome float64) float64 {
	if income <= 0 && yearlyIncome <= 0 {
		return 40 // new user – neutral
	}
	if income <= 0 {
		return 20
	}
	score := 70.0
	if prevIncome > 0 {
		change := (income - prevIncome) / prevIncome
		switch {
		case change >= 0.10:
			score = 95
		case change >= 0:
			score = 80
		case change >= -0.10:
			score = 65
		default:
			score = 45
		}
	}
	return score
}

func calcSavingsScore(income, expense float64) float64 {
	if income <= 0 {
		return 50
	}
	rate := savingsRate(income, expense)
	switch {
	case rate >= 30:
		return 100
	case rate >= 20:
		return 85
	case rate >= 10:
		return 70
	case rate >= 0:
		return 50
	default: // spending more than earning
		return 10
	}
}

func calcExpenseRatio(income, expense float64) float64 {
	if income <= 0 {
		return 50
	}
	ratio := (expense / income) * 100
	return clamp(ratio, 0, 100)
}

func calcBudgetHealth(ctx context.Context, uc *AnalyticsUseCase, userID string, budgets []domain.Budget, from, to time.Time) float64 {
	active := 0
	onTrack := 0
	for _, b := range budgets {
		if !b.IsActive {
			continue
		}
		active++
		spends, err := uc.expenses.SumByCategory(ctx, userID, from, to)
		if err != nil {
			continue
		}
		var spent float64
		for _, cs := range spends {
			if cs.Category == string(b.Category) {
				spent = cs.Amount
				break
			}
		}
		if b.Amount > 0 && spent/b.Amount < 1.0 {
			onTrack++
		}
	}
	if active == 0 {
		return 70 // no budgets set – neutral
	}
	return (float64(onTrack) / float64(active)) * 100
}

func calcSubscriptionHealth(income float64, subs []domain.Subscription) float64 {
	if income <= 0 {
		return 80
	}
	var monthly float64
	for _, s := range subs {
		if !s.IsActive {
			continue
		}
		switch s.BillingCycle {
		case "monthly":
			monthly += s.Amount
		case "yearly":
			monthly += s.Amount / 12
		case "weekly":
			monthly += s.Amount * 4
		}
	}
	ratio := (monthly / income) * 100
	switch {
	case ratio <= 5:
		return 100
	case ratio <= 10:
		return 80
	case ratio <= 20:
		return 60
	default:
		return 30
	}
}

func calcDebtHealth(expense, prevExpense, income float64) float64 {
	if income <= 0 {
		return 60
	}
	// Proxy: expense growth trend vs income
	if prevExpense <= 0 {
		return 70
	}
	growth := (expense - prevExpense) / prevExpense
	switch {
	case growth <= 0:
		return 90
	case growth <= 0.05:
		return 75
	case growth <= 0.15:
		return 55
	default:
		return 30
	}
}

func buildInsights(income, expense, savingsScore, budgetHealth, subscriptionHealth float64) []string {
	var out []string
	if income <= 0 {
		out = append(out, "Add your income sources to get a complete financial picture.")
	}
	rate := savingsRate(income, expense)
	switch {
	case rate >= 20:
		out = append(out, "Great job! You're saving more than 20% of your income.")
	case rate >= 10:
		out = append(out, "You're saving around 10% of income. Aim for 20% for a stronger safety net.")
	case rate > 0:
		out = append(out, "Your savings rate is low. Consider reducing discretionary expenses.")
	default:
		out = append(out, "You're spending more than you earn this month. Review your budget immediately.")
	}
	if budgetHealth < 50 {
		out = append(out, "More than half your budgets are over-limit. Set tighter spending controls.")
	} else if budgetHealth < 80 {
		out = append(out, "Some budgets are approaching their limits. Monitor spending closely.")
	}
	if subscriptionHealth < 60 {
		out = append(out, "Subscription costs are high relative to your income. Consider cancelling unused services.")
	}
	return out
}

func savingsRate(income, expense float64) float64 {
	if income <= 0 {
		return 0
	}
	return ((income - expense) / income) * 100
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}

func marshalInsights(insights []string) ([]byte, error) {
	return json.Marshal(insights)
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
