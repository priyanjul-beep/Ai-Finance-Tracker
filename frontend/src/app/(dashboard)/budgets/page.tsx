"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Plus, Loader2, Trash2, AlertTriangle, CheckCircle, XCircle,
  ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight,
} from "lucide-react";
import { useBudgets, useDeleteBudget } from "@/hooks/useExpenses";
import { formatCurrency } from "@/lib/utils";

const MONTHS = [
  "January","February","March","April","May","June",
  "July","August","September","October","November","December",
];

const currentYear  = new Date().getFullYear();
const currentMonth = new Date().getMonth() + 1;

const PAGE_SIZE = 6;

const BUDGET_CATEGORIES = [
  "Food & Dining","Transportation","Shopping","Entertainment",
  "Healthcare","Housing","Utilities","Travel","Education",
  "Personal Care","Subscriptions","Other",
];

const STATUS_CONFIG = {
  "on-track":    { icon: CheckCircle,    color: "text-green-600",  bg: "bg-green-100",  bar: "bg-green-500",  label: "On Track"    },
  "warning":     { icon: AlertTriangle,  color: "text-yellow-600", bg: "bg-yellow-100", bar: "bg-yellow-500", label: "Warning"     },
  "over-budget": { icon: XCircle,        color: "text-red-600",    bg: "bg-red-100",    bar: "bg-red-500",    label: "Over Budget" },
};

function pageWindow(current: number, total: number): (number | "\u2026")[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
  const pages: (number | "\u2026")[] = [1];
  if (current > 3) pages.push("\u2026");
  for (let p = Math.max(2, current - 1); p <= Math.min(total - 1, current + 1); p++) pages.push(p);
  if (current < total - 2) pages.push("\u2026");
  pages.push(total);
  return pages;
}

export default function BudgetsPage() {
  const [selectedYear,  setSelectedYear]  = useState(currentYear);
  const [selectedMonth, setSelectedMonth] = useState(currentMonth);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [page, setPage] = useState(1);
  const [categoryFilter, setCategoryFilter] = useState("");

  const { data: budgets = [], isLoading } = useBudgets(selectedYear, selectedMonth, categoryFilter);
  const { mutate: deleteBudget, isPending: isDeleting } = useDeleteBudget();

  const totalPages   = Math.max(1, Math.ceil(budgets.length / PAGE_SIZE));
  const pagedBudgets = budgets.slice((page - 1) * PAGE_SIZE, page * PAGE_SIZE);
  const rangeStart   = budgets.length === 0 ? 0 : (page - 1) * PAGE_SIZE + 1;
  const rangeEnd     = Math.min(page * PAGE_SIZE, budgets.length);

  const totalBudgeted = budgets.reduce((s, b) => s + b.amount, 0);
  const totalSpent    = budgets.reduce((s, b) => s + (b.spent ?? 0), 0);
  const currency      = budgets[0]?.currency ?? "INR";

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Budgets</h1>
          <p className="text-sm text-muted-foreground">
            {budgets.length} budget{budgets.length !== 1 ? "s" : ""} for {MONTHS[selectedMonth - 1]} {selectedYear}
          </p>
        </div>
        <Link
          href="/budgets/new"
          className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="h-4 w-4" />
          Add Budget
        </Link>
      </div>

      {/* Month / Year selector + Search */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:flex-wrap">
        <select
          value={selectedMonth}
          onChange={(e) => { setSelectedMonth(Number(e.target.value)); setPage(1); }}
          className="rounded-lg border border-input bg-background px-3 py-1.5 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
        >
          {MONTHS.map((m, i) => (
            <option key={m} value={i + 1}>{m}</option>
          ))}
        </select>
        <select
          value={selectedYear}
          onChange={(e) => { setSelectedYear(Number(e.target.value)); setPage(1); }}
          className="rounded-lg border border-input bg-background px-3 py-1.5 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
        >
          {[currentYear - 1, currentYear, currentYear + 1].map((y) => (
            <option key={y} value={y}>{y}</option>
          ))}
        </select>

        {/* Category filter */}
        <select
          value={categoryFilter}
          onChange={(e) => { setCategoryFilter(e.target.value); setPage(1); }}
          className="rounded-lg border border-input bg-background px-3 py-1.5 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary min-w-[160px]"
        >
          <option value="">All Categories</option>
          {BUDGET_CATEGORIES.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
      </div>

      {/* Summary row */}
      {budgets.length > 0 && (
        <div className="grid grid-cols-3 gap-4">
          {[
            { label: "Total Budgeted", value: formatCurrency(totalBudgeted, currency), color: "text-foreground" },
            { label: "Total Spent",    value: formatCurrency(totalSpent,    currency), color: "text-destructive" },
            { label: "Remaining",      value: formatCurrency(totalBudgeted - totalSpent, currency), color: totalBudgeted - totalSpent >= 0 ? "text-green-600" : "text-destructive" },
          ].map(({ label, value, color }) => (
            <div key={label} className="rounded-xl border border-border bg-card p-4 shadow-sm">
              <p className="text-xs text-muted-foreground">{label}</p>
              <p className={`mt-1 text-lg font-bold ${color}`}>{value}</p>
            </div>
          ))}
        </div>
      )}

      {/* Budget cards */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : budgets.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-border bg-card py-20 text-center shadow-sm">
          <p className="text-sm font-medium text-muted-foreground">
            No budgets for {MONTHS[selectedMonth - 1]} {selectedYear}
          </p>
          <Link href="/budgets/new" className="mt-3 text-sm font-medium text-primary hover:underline">
            Create your first budget →
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {pagedBudgets.map((budget) => {
            const status = (budget.status ?? "on-track") as keyof typeof STATUS_CONFIG;
            const cfg    = STATUS_CONFIG[status] ?? STATUS_CONFIG["on-track"];
            const Icon   = cfg.icon;
            const pct    = Math.min(budget.percent ?? 0, 100);

            return (
              <div key={budget.id} className="rounded-xl border border-border bg-card p-5 shadow-sm space-y-4">
                {/* Top row */}
                <div className="flex items-start justify-between">
                  <div>
                    <p className="font-semibold">{budget.category}</p>
                    {budget.description && (
                      <p className="text-xs text-muted-foreground mt-0.5">{budget.description}</p>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className={`flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${cfg.bg} ${cfg.color}`}>
                      <Icon className="h-3 w-3" />
                      {cfg.label}
                    </span>
                    <button
                      onClick={() => setConfirmDeleteId(budget.id)}
                      className="rounded p-1 hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>

                {/* Progress bar */}
                <div>
                  <div className="mb-1.5 flex items-center justify-between text-xs">
                    <span className="text-muted-foreground">
                      {formatCurrency(budget.spent ?? 0, budget.currency)} spent
                    </span>
                    <span className="font-medium">{pct.toFixed(0)}%</span>
                  </div>
                  <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${cfg.bar}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>

                {/* Amount row */}
                <div className="flex items-center justify-between text-sm">
                  <span className="text-muted-foreground">
                    Remaining: <span className={budget.remaining >= 0 ? "text-green-600 font-medium" : "text-destructive font-medium"}>
                      {formatCurrency(budget.remaining ?? 0, budget.currency)}
                    </span>
                  </span>
                  <span className="font-semibold">
                    {formatCurrency(budget.amount, budget.currency)}
                  </span>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Pagination */}
      {budgets.length > 0 && (
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between text-sm">
          <p className="text-muted-foreground">
            Showing <span className="font-medium text-foreground">{rangeStart}–{rangeEnd}</span> of{" "}
            <span className="font-medium text-foreground">{budgets.length}</span> budgets
          </p>
          <div className="flex items-center gap-1">
            <button onClick={() => setPage(1)} disabled={page === 1} aria-label="First page"
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronsLeft className="h-4 w-4" />
            </button>
            <button onClick={() => setPage((p) => Math.max(1, p - 1))} disabled={page === 1} aria-label="Previous page"
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronLeft className="h-4 w-4" />
            </button>
            {pageWindow(page, totalPages).map((p, i) =>
              p === "…" ? (
                <span key={`e-${i}`} className="flex h-8 w-8 items-center justify-center text-muted-foreground select-none">…</span>
              ) : (
                <button key={p} onClick={() => setPage(p)}
                  className={`flex h-8 min-w-[2rem] items-center justify-center rounded-lg border px-2 transition-colors font-medium ${
                    p === page ? "border-primary bg-primary text-primary-foreground" : "border-border hover:bg-muted text-foreground"
                  }`}>
                  {p}
                </button>
              )
            )}
            <button onClick={() => setPage((p) => Math.min(totalPages, p + 1))} disabled={page === totalPages} aria-label="Next page"
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronRight className="h-4 w-4" />
            </button>
            <button onClick={() => setPage(totalPages)} disabled={page === totalPages} aria-label="Last page"
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              <ChevronsRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}

      {/* Delete confirmation modal */}
      {confirmDeleteId && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
          onClick={() => setConfirmDeleteId(null)}
        >
          <div
            className="mx-4 w-full max-w-sm rounded-xl border border-border bg-background p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-base font-semibold">Delete budget?</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              This action cannot be undone.
            </p>
            <div className="mt-6 flex gap-3">
              <button
                onClick={() => setConfirmDeleteId(null)}
                className="flex-1 rounded-lg border border-border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
              >
                Cancel
              </button>
              <button
                disabled={isDeleting}
                onClick={() => deleteBudget(confirmDeleteId, { onSuccess: () => setConfirmDeleteId(null) })}
                className="flex flex-1 items-center justify-center gap-2 rounded-lg bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground hover:bg-destructive/90 disabled:opacity-60 transition-colors"
              >
                {isDeleting && <Loader2 className="h-4 w-4 animate-spin" />}
                Delete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
