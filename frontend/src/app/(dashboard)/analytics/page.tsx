"use client";

import { useState } from "react";
import { motion } from "framer-motion";
import {
  TrendingUp,
  TrendingDown,
  Sparkles,
  AlertCircle,
  Activity,
  PiggyBank,
  Loader2,
} from "lucide-react";
import {
  useDashboard,
  useMonthlyReport,
  usePredictions,
  useInsights,
  useHealthScore,
} from "@/hooks/useDashboard";
import {
  CategoryPieChart,
  SpendingBarChart,
  IncomeExpenseChart,
} from "@/components/charts";
import { formatCurrency, formatPercent } from "@/lib/utils";

const MONTHS = [
  "January", "February", "March", "April", "May", "June",
  "July", "August", "September", "October", "November", "December",
];

const currentYear = new Date().getFullYear();
const currentMonth = new Date().getMonth() + 1;

function ScoreRing({ score }: { score: number }) {
  const color =
    score >= 75 ? "#10b981" : score >= 50 ? "#f59e0b" : "#ef4444";
  const radius = 40;
  const circumference = 2 * Math.PI * radius;
  const dash = (score / 100) * circumference;

  return (
    <svg width={100} height={100} viewBox="0 0 100 100">
      <circle cx="50" cy="50" r={radius} fill="none" stroke="hsl(var(--border))" strokeWidth={10} />
      <circle
        cx="50" cy="50" r={radius} fill="none"
        stroke={color} strokeWidth={10}
        strokeDasharray={`${dash} ${circumference}`}
        strokeLinecap="round"
        transform="rotate(-90 50 50)"
        style={{ transition: "stroke-dasharray 0.6s ease" }}
      />
      <text x="50" y="54" textAnchor="middle" fontSize="18" fontWeight="bold" fill={color}>
        {score}
      </text>
    </svg>
  );
}

function StatTile({
  label,
  value,
  sub,
  icon,
  loading,
}: {
  label: string;
  value: string;
  sub?: string;
  icon: React.ReactNode;
  loading?: boolean;
}) {
  return (
    <div className="rounded-xl border border-border bg-card p-4 shadow-sm flex items-center gap-4">
      <div className="flex h-10 w-10 flex-shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
        {icon}
      </div>
      <div className="min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        {loading ? (
          <div className="mt-1 h-5 w-24 animate-pulse rounded bg-muted" />
        ) : (
          <p className="text-lg font-bold truncate">{value}</p>
        )}
        {sub && <p className="text-xs text-muted-foreground">{sub}</p>}
      </div>
    </div>
  );
}

export default function AnalyticsPage() {
  const [selectedMonth, setSelectedMonth] = useState(currentMonth);
  const [selectedYear, setSelectedYear] = useState(currentYear);

  const { data: dashboard, isLoading: dashLoading } = useDashboard();
  const { data: monthly, isLoading: monthlyLoading } = useMonthlyReport(selectedYear, selectedMonth);
  const { data: predictions, isLoading: predictLoading } = usePredictions();
  const { data: insights } = useInsights();
  const { data: health, isLoading: healthLoading } = useHealthScore();

  const currency = "INR";

  // Build a simple time-series from monthly report for area chart
  const timeSeriesData = monthly
    ? [{ date: `${MONTHS[monthly.month - 1].slice(0, 3)} ${monthly.year}`, income: monthly.total_income, expenses: monthly.total_expenses, savings: monthly.net_savings }]
    : [];

  return (
    <div className="space-y-6">
      {/* Title */}
      <div>
        <h1 className="text-2xl font-bold">Analytics</h1>
        <p className="text-sm text-muted-foreground">
          Deep insights into your financial health
        </p>
      </div>

      {/* Summary stat tiles */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatTile
          label="Monthly Spend"
          value={formatCurrency(dashboard?.monthly_spend ?? 0, currency)}
          icon={<TrendingDown className="h-5 w-5" />}
          loading={dashLoading}
        />
        <StatTile
          label="Monthly Income"
          value={formatCurrency(dashboard?.total_income ?? 0, currency)}
          icon={<TrendingUp className="h-5 w-5" />}
          loading={dashLoading}
        />
        <StatTile
          label="Savings Rate"
          value={formatPercent(dashboard?.savings_rate ?? 0)}
          icon={<PiggyBank className="h-5 w-5" />}
          loading={dashLoading}
        />
        <StatTile
          label="Weekly Spend"
          value={formatCurrency(dashboard?.weekly_spend ?? 0, currency)}
          icon={<Activity className="h-5 w-5" />}
          loading={dashLoading}
        />
      </div>

      {/* Month / Year selector */}
      <div className="flex items-center gap-3">
        <select
          value={selectedMonth}
          onChange={(e) => setSelectedMonth(Number(e.target.value))}
          className="rounded-lg border border-input bg-background px-3 py-1.5 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
        >
          {MONTHS.map((m, i) => (
            <option key={m} value={i + 1}>{m}</option>
          ))}
        </select>
        <select
          value={selectedYear}
          onChange={(e) => setSelectedYear(Number(e.target.value))}
          className="rounded-lg border border-input bg-background px-3 py-1.5 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
        >
          {[currentYear - 1, currentYear, currentYear + 1].map((y) => (
            <option key={y} value={y}>{y}</option>
          ))}
        </select>
      </div>

      {/* Charts row */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Category breakdown */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <h2 className="mb-4 text-sm font-semibold">Spend by Category</h2>
          {dashLoading ? (
            <div className="flex h-64 items-center justify-center">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : (dashboard?.spend_by_category?.length ?? 0) > 0 ? (
            <CategoryPieChart data={dashboard!.spend_by_category} currency={currency} />
          ) : (
            <p className="flex h-64 items-center justify-center text-sm text-muted-foreground">No data yet</p>
          )}
        </div>

        {/* Bar chart top categories */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <h2 className="mb-4 text-sm font-semibold">Top Categories</h2>
          {dashLoading ? (
            <div className="flex h-64 items-center justify-center">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : (dashboard?.spend_by_category?.length ?? 0) > 0 ? (
            <SpendingBarChart data={dashboard!.spend_by_category} currency={currency} />
          ) : (
            <p className="flex h-64 items-center justify-center text-sm text-muted-foreground">No data yet</p>
          )}
        </div>
      </div>

      {/* Monthly report + Income/Expense trend */}
      {timeSeriesData.length > 0 && (
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <h2 className="mb-4 text-sm font-semibold">
            Income vs Expenses — {MONTHS[selectedMonth - 1]} {selectedYear}
          </h2>
          <IncomeExpenseChart data={timeSeriesData} currency={currency} />
        </div>
      )}

      {/* Monthly summary cards */}
      {monthly && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          {[
            { label: "Total Income", value: formatCurrency(monthly.total_income, currency) },
            { label: "Total Expenses", value: formatCurrency(monthly.total_expenses, currency) },
            { label: "Net Savings", value: formatCurrency(monthly.net_savings, currency) },
            { label: "Savings Rate", value: formatPercent(monthly.savings_rate) },
          ].map(({ label, value }) => (
            <div key={label} className="rounded-xl border border-border bg-card p-4 shadow-sm">
              <p className="text-xs text-muted-foreground">{label}</p>
              <p className="mt-1 text-lg font-bold">{value}</p>
            </div>
          ))}
        </div>
      )}

      {/* AI predictions + health score row */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Predictions */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <Sparkles className="h-4 w-4 text-primary" />
            <h2 className="text-sm font-semibold">AI Spending Prediction</h2>
          </div>
          {predictLoading ? (
            <div className="space-y-2">
              {[1, 2, 3].map((i) => <div key={i} className="h-4 animate-pulse rounded bg-muted" />)}
            </div>
          ) : predictions ? (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Next month estimate</span>
                <span className="font-semibold">{formatCurrency(predictions.next_month_prediction, currency)}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Confidence</span>
                <span className="font-medium">{Math.round(predictions.confidence * 100)}%</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Trend</span>
                <span className={`flex items-center gap-1 text-sm font-medium ${
                  predictions.trend === "decreasing" ? "text-green-600" :
                  predictions.trend === "increasing" ? "text-red-500" : "text-yellow-500"
                }`}>
                  {predictions.trend === "increasing" ? <TrendingUp className="h-3.5 w-3.5" /> : <TrendingDown className="h-3.5 w-3.5" />}
                  {predictions.trend}
                </span>
              </div>
              {predictions.message && (
                <p className="rounded-lg bg-muted/50 p-3 text-xs text-muted-foreground">
                  {predictions.message}
                </p>
              )}
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No predictions available</p>
          )}
        </div>

        {/* Health score */}
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <Activity className="h-4 w-4 text-primary" />
            <h2 className="text-sm font-semibold">Financial Health Score</h2>
          </div>
          {healthLoading ? (
            <div className="flex h-32 items-center justify-center">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
            </div>
          ) : health ? (
            <div className="flex items-center gap-6">
              <ScoreRing score={health.score} />
              <div className="space-y-2 flex-1 text-sm">
                {[
                  { label: "Income", val: health.income_score },
                  { label: "Savings", val: health.savings_score },
                  { label: "Budget", val: health.budget_health },
                  { label: "Subscriptions", val: health.subscription_health },
                ].map(({ label, val }) => (
                  <div key={label} className="flex items-center gap-2">
                    <span className="w-24 text-muted-foreground text-xs">{label}</span>
                    <div className="flex-1 h-1.5 rounded-full bg-muted overflow-hidden">
                      <div
                        className="h-full rounded-full bg-primary transition-all duration-500"
                        style={{ width: `${Math.min(val, 100)}%` }}
                      />
                    </div>
                    <span className="text-xs font-medium w-6 text-right">{val}</span>
                  </div>
                ))}
              </div>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">No health data available</p>
          )}
        </div>
      </div>

      {/* AI Insights */}
      {insights && insights.length > 0 && (
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <Sparkles className="h-4 w-4 text-primary" />
            <h2 className="text-sm font-semibold">AI Insights</h2>
          </div>
          <ul className="space-y-2">
            {insights.map((insight, i) => (
              <motion.li
                key={i}
                initial={{ opacity: 0, x: -8 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ delay: i * 0.05 }}
                className="flex items-start gap-2 rounded-lg bg-muted/40 px-3 py-2.5 text-sm"
              >
                <AlertCircle className="mt-0.5 h-3.5 w-3.5 flex-shrink-0 text-primary" />
                {insight}
              </motion.li>
            ))}
          </ul>
        </div>
      )}

      {/* AI recommendations from monthly report */}
      {monthly?.ai_recommendations && monthly.ai_recommendations.length > 0 && (
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-4">
            <Sparkles className="h-4 w-4 text-primary" />
            <h2 className="text-sm font-semibold">
              Recommendations for {MONTHS[selectedMonth - 1]}
            </h2>
          </div>
          <ul className="space-y-2">
            {monthly.ai_recommendations.map((rec, i) => (
              <li key={i} className="flex items-start gap-2 text-sm text-muted-foreground">
                <span className="mt-0.5 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-primary" />
                {rec}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
}
