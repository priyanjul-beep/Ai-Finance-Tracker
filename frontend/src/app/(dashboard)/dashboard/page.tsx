"use client";

import { Wallet, TrendingUp, TrendingDown, PiggyBank, Sparkles, AlertCircle } from "lucide-react";
import { motion } from "framer-motion";
import { StatCard } from "@/components/dashboard/StatCard";
import { CategoryPieChart, SpendingBarChart } from "@/components/charts";
import { useDashboard, useInsights } from "@/hooks/useDashboard";
import {
  formatCurrency,
  formatPercent,
  getCategoryIcon,
  formatDate,
} from "@/lib/utils";

export default function DashboardPage() {
  const { data: dashboard, isLoading } = useDashboard();
  const { data: insights } = useInsights();

  return (
    <div className="space-y-6">
      {/* Page title */}
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Your financial overview at a glance
        </p>
      </div>

      {/* Stat cards */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard
          title="Total Balance"
          value={dashboard?.total_balance ?? 0}
          change={2.4}
          trend="up"
          icon={<Wallet className="h-5 w-5" />}
          loading={isLoading}
        />
        <StatCard
          title="Monthly Income"
          value={dashboard?.total_income ?? 0}
          change={5.1}
          trend="up"
          icon={<TrendingUp className="h-5 w-5" />}
          loading={isLoading}
        />
        <StatCard
          title="Monthly Expenses"
          value={dashboard?.monthly_spend ?? 0}
          change={-3.2}
          trend="down"
          icon={<TrendingDown className="h-5 w-5" />}
          loading={isLoading}
        />
        <StatCard
          title="Savings Rate"
          value={dashboard?.savings_rate ?? 0}
          description="of income saved this month"
          icon={<PiggyBank className="h-5 w-5" />}
          loading={isLoading}
        />
      </div>

      {/* Charts row */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Category pie */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="rounded-xl border border-border bg-card p-5"
        >
          <h3 className="mb-4 text-sm font-semibold">Spend by Category</h3>
          {dashboard?.spend_by_category && (
            <CategoryPieChart data={dashboard.spend_by_category} />
          )}
          {isLoading && (
            <div className="h-[280px] animate-pulse rounded bg-muted" />
          )}
        </motion.div>

        {/* Spending bar */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.15 }}
          className="rounded-xl border border-border bg-card p-5"
        >
          <h3 className="mb-4 text-sm font-semibold">Top Categories</h3>
          {dashboard?.spend_by_category && (
            <SpendingBarChart data={dashboard.spend_by_category} />
          )}
          {isLoading && (
            <div className="h-[280px] animate-pulse rounded bg-muted" />
          )}
        </motion.div>
      </div>

      {/* Bottom row: recent expenses + AI insights */}
      <div className="grid gap-6 lg:grid-cols-2">
        {/* Recent expenses */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="rounded-xl border border-border bg-card p-5"
        >
          <h3 className="mb-4 text-sm font-semibold">Recent Expenses</h3>
          {isLoading && (
            <div className="space-y-3">
              {[1, 2, 3, 4, 5].map((i) => (
                <div
                  key={i}
                  className="flex items-center gap-3"
                >
                  <div className="h-9 w-9 animate-pulse rounded-lg bg-muted flex-shrink-0" />
                  <div className="flex-1 space-y-1.5">
                    <div className="h-3.5 w-32 animate-pulse rounded bg-muted" />
                    <div className="h-3 w-20 animate-pulse rounded bg-muted" />
                  </div>
                  <div className="h-4 w-16 animate-pulse rounded bg-muted" />
                </div>
              ))}
            </div>
          )}
          {dashboard?.recent_expenses && (
            <ul className="space-y-3">
              {dashboard.recent_expenses.slice(0, 6).map((exp) => (
                <li key={exp.id} className="flex items-center gap-3">
                  <div className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-lg bg-muted text-base">
                    {getCategoryIcon(exp.category)}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">
                      {exp.merchant || exp.description}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {formatDate(exp.date)}
                    </p>
                  </div>
                  <span className="text-sm font-semibold tabular-nums text-destructive">
                    -{formatCurrency(exp.amount)}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </motion.div>

        {/* AI insights */}
        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.25 }}
          className="rounded-xl border border-border bg-card p-5"
        >
          <div className="mb-4 flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-primary" />
            <h3 className="text-sm font-semibold">AI Insights</h3>
          </div>
          {insights && insights.length > 0 ? (
            <ul className="space-y-3">
              {insights.map((insight, i) => (
                <li
                  key={i}
                  className="flex items-start gap-2.5 rounded-lg bg-primary/5 p-3 text-sm"
                >
                  <AlertCircle className="mt-0.5 h-4 w-4 flex-shrink-0 text-primary" />
                  <span>{insight}</span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-muted-foreground">
              Keep tracking your expenses to unlock AI insights.
            </p>
          )}

          {/* Predictions teaser */}
          {dashboard?.predictions && (
            <div className="mt-4 rounded-lg border border-primary/20 bg-primary/5 p-4">
              <p className="text-xs font-semibold text-primary uppercase tracking-wide mb-1">
                Next Month Prediction
              </p>
              <p className="text-xl font-bold">
                {formatCurrency(dashboard.predictions.next_month_prediction)}
              </p>
              <p className="text-xs text-muted-foreground mt-0.5">
                {formatPercent(dashboard.predictions.confidence)} confidence ·{" "}
                {dashboard.predictions.trend} trend
              </p>
            </div>
          )}
        </motion.div>
      </div>
    </div>
  );
}
