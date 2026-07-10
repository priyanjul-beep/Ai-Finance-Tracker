"use client";

import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  AreaChart,
  Area,
} from "recharts";
import { getCategoryColor, formatCurrency } from "@/lib/utils";
import type { CategorySpend, TimeSeriesPoint } from "@/types";

// ─── Category Pie Chart ───────────────────────────────────────────────────────

interface CategoryPieChartProps {
  data: CategorySpend[];
  currency?: string;
}

export function CategoryPieChart({
  data,
  currency = "INR",
}: CategoryPieChartProps) {
  const chartData = data.map((d) => ({
    name: d.category.charAt(0).toUpperCase() + d.category.slice(1),
    value: d.amount,
    color: getCategoryColor(d.category),
  }));

  return (
    <ResponsiveContainer width="100%" height={280}>
      <PieChart>
        <Pie
          data={chartData}
          cx="50%"
          cy="50%"
          outerRadius={90}
          innerRadius={50}
          paddingAngle={3}
          dataKey="value"
        >
          {chartData.map((entry, index) => (
            <Cell key={index} fill={entry.color} />
          ))}
        </Pie>
        <Tooltip
          formatter={(value: number) => formatCurrency(value, currency)}
        />
        <Legend
          iconType="circle"
          iconSize={8}
          formatter={(value) => (
            <span className="text-xs text-foreground">{value}</span>
          )}
        />
      </PieChart>
    </ResponsiveContainer>
  );
}

// ─── Spending Bar Chart ───────────────────────────────────────────────────────

interface SpendingBarChartProps {
  data: CategorySpend[];
  currency?: string;
}

export function SpendingBarChart({
  data,
  currency = "INR",
}: SpendingBarChartProps) {
  const chartData = data
    .sort((a, b) => b.amount - a.amount)
    .slice(0, 8)
    .map((d) => ({
      category: d.category.charAt(0).toUpperCase() + d.category.slice(1),
      amount: d.amount,
      fill: getCategoryColor(d.category),
    }));

  return (
    <ResponsiveContainer width="100%" height={280}>
      <BarChart data={chartData} margin={{ left: 10 }}>
        <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
        <XAxis
          dataKey="category"
          tick={{ fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          tickFormatter={(v) => v >= 1000 ? `₹${(v / 1000).toFixed(1)}k` : `₹${v}`}
          tick={{ fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <Tooltip formatter={(v: number) => formatCurrency(v, currency)} />
        <Bar dataKey="amount" radius={[4, 4, 0, 0]}>
          {chartData.map((entry, index) => (
            <Cell key={index} fill={entry.fill} />
          ))}
        </Bar>
      </BarChart>
    </ResponsiveContainer>
  );
}

// ─── Income vs Expense Area Chart ─────────────────────────────────────────────

interface IncomeExpenseChartProps {
  data: TimeSeriesPoint[];
  currency?: string;
}

export function IncomeExpenseChart({
  data,
  currency = "INR",
}: IncomeExpenseChartProps) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <AreaChart data={data} margin={{ left: 10, right: 10 }}>
        <defs>
          <linearGradient id="incomeGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
            <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
          </linearGradient>
          <linearGradient id="expenseGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="#ef4444" stopOpacity={0.3} />
            <stop offset="95%" stopColor="#ef4444" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="hsl(var(--border))" />
        <XAxis
          dataKey="date"
          tick={{ fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <YAxis
          tickFormatter={(v) => v >= 1000 ? `₹${(v / 1000).toFixed(1)}k` : `₹${v}`}
          tick={{ fontSize: 11 }}
          tickLine={false}
          axisLine={false}
        />
        <Tooltip formatter={(v: number) => formatCurrency(v, currency)} />
        <Legend iconType="circle" iconSize={8} />
        <Area
          type="monotone"
          dataKey="income"
          stroke="#10b981"
          fill="url(#incomeGradient)"
          strokeWidth={2}
          name="Income"
        />
        <Area
          type="monotone"
          dataKey="expenses"
          stroke="#ef4444"
          fill="url(#expenseGradient)"
          strokeWidth={2}
          name="Expenses"
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
