"use client";

import { useState, useRef } from "react";
import Link from "next/link";
import {
  Plus,
  Loader2,
  Trash2,
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
  Search,
  X,
} from "lucide-react";
import { useIncomeList, useDeleteIncome } from "@/hooks/useExpenses";
import { formatCurrency, formatDate } from "@/lib/utils";

const INCOME_CATEGORIES = [
  "Salary", "Freelance", "Business", "Investment",
  "Rental", "Interest", "Dividend", "Gift", "Other",
];

const CATEGORY_COLORS: Record<string, string> = {
  Salary: "bg-green-100 text-green-700",
  Freelance: "bg-blue-100 text-blue-700",
  Business: "bg-purple-100 text-purple-700",
  Investment: "bg-yellow-100 text-yellow-700",
  Rental: "bg-orange-100 text-orange-700",
  Interest: "bg-cyan-100 text-cyan-700",
  Dividend: "bg-teal-100 text-teal-700",
  Gift: "bg-pink-100 text-pink-700",
  Other: "bg-gray-100 text-gray-700",
};

const LIMIT = 10;

function pageWindow(current: number, total: number): (number | "\u2026")[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1);
  const pages: (number | "\u2026")[] = [1];
  if (current > 3) pages.push("\u2026");
  for (let p = Math.max(2, current - 1); p <= Math.min(total - 1, current + 1); p++) pages.push(p);
  if (current < total - 2) pages.push("\u2026");
  pages.push(total);
  return pages;
}

export default function IncomePage() {
  const [page, setPage] = useState(1);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  // search state (committed = sent to API)
  const [sourceInput,    setSourceInput]    = useState("");
  const [categoryInput,  setCategoryInput]  = useState("");
  const [sourceFilter,   setSourceFilter]   = useState("");
  const [categoryFilter, setCategoryFilter] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  const applySearch = () => {
    setSourceFilter(sourceInput);
    setCategoryFilter(categoryInput);
    setPage(1);
  };

  const clearSearch = () => {
    setSourceInput(""); setCategoryInput("");
    setSourceFilter(""); setCategoryFilter("");
    setPage(1);
  };

  const hasFilter = sourceFilter !== "" || categoryFilter !== "";

  const { data, isLoading } = useIncomeList({
    page, limit: LIMIT,
    source:   sourceFilter   || undefined,
    category: categoryFilter || undefined,
  });
  const { mutate: deleteIncome, isPending: isDeleting } = useDeleteIncome();

  const incomes    = data?.data        ?? [];
  const total      = data?.total       ?? 0;
  const totalPages = data?.total_pages ?? 1;

  const rangeStart = total === 0 ? 0 : (page - 1) * LIMIT + 1;
  const rangeEnd   = Math.min(page * LIMIT, total);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Income</h1>
          <p className="text-sm text-muted-foreground">{total} total entries</p>
        </div>
        <Link
          href="/income/new"
          className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="h-4 w-4" />
          Add Income
        </Link>
      </div>

      {/* Search bar */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center">
        {/* Source text input */}
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <input
            ref={inputRef}
            type="text"
            placeholder="Search by source…"
            value={sourceInput}
            onChange={(e) => setSourceInput(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && applySearch()}
            className="w-full rounded-lg border border-input bg-background py-2 pl-9 pr-4 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
          />
        </div>
        {/* Category dropdown */}
        <select
          value={categoryInput}
          onChange={(e) => setCategoryInput(e.target.value)}
          className="rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary min-w-[160px]"
        >
          <option value="">All Categories</option>
          {INCOME_CATEGORIES.map((c) => (
            <option key={c} value={c}>{c}</option>
          ))}
        </select>
        {/* Search button */}
        <button
          onClick={applySearch}
          className="flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Search className="h-4 w-4" />
          Search
        </button>
        {/* Clear button */}
        {hasFilter && (
          <button
            onClick={clearSearch}
            className="flex items-center gap-1.5 rounded-lg border border-border px-3 py-2 text-sm text-muted-foreground hover:bg-muted transition-colors"
          >
            <X className="h-4 w-4" />
            Clear
          </button>
        )}
      </div>

      {/* Table */}
      <div className="rounded-xl border border-border bg-card shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : incomes.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <p className="text-sm font-medium text-muted-foreground">
              No income entries yet
            </p>
            <Link
              href="/income/new"
              className="mt-3 text-sm font-medium text-primary hover:underline"
            >
              Add your first income →
            </Link>
          </div>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-border bg-muted/40">
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Date</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Source</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Category</th>
                <th className="px-4 py-3 text-left font-medium text-muted-foreground">Description</th>
                <th className="px-4 py-3 text-right font-medium text-muted-foreground">Amount</th>
                <th className="px-4 py-3 text-right font-medium text-muted-foreground">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {incomes.map((income) => (
                <tr key={income.id} className="hover:bg-muted/30 transition-colors">
                  <td className="px-4 py-3 text-muted-foreground whitespace-nowrap">
                    {formatDate(new Date(income.date))}
                  </td>
                  <td className="px-4 py-3 font-medium">{income.source}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        CATEGORY_COLORS[income.category] ?? "bg-gray-100 text-gray-700"
                      }`}
                    >
                      {income.category || "Other"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-muted-foreground max-w-[200px] truncate">
                    {income.description || "—"}
                  </td>
                  <td className="px-4 py-3 text-right font-semibold text-green-600">
                    +{formatCurrency(income.amount, income.currency)}
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center justify-end">
                      <button
                        onClick={() => setConfirmDeleteId(income.id)}
                        className="rounded p-1.5 hover:bg-destructive/10 transition-colors text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Pagination */}
      {total > 0 && (
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between text-sm">
          <p className="text-muted-foreground">
            Showing <span className="font-medium text-foreground">{rangeStart}–{rangeEnd}</span> of{" "}
            <span className="font-medium text-foreground">{total}</span> entries
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
            <h2 className="text-base font-semibold">Delete income entry?</h2>
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
                onClick={() =>
                  deleteIncome(confirmDeleteId!, {
                    onSuccess: () => setConfirmDeleteId(null),
                  })
                }
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
