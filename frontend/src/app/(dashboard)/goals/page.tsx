"use client";

import { useState } from "react";
import Link from "next/link";
import {
  Plus, Loader2, Trash2, Target, CheckCircle,
  PauseCircle, TrendingUp, Calendar,
} from "lucide-react";
import { useGoals, useCreateGoal, useContributeToGoal } from "@/hooks/useExpenses";
import { goalService } from "@/services/expense.service";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-client";
import { formatCurrency, formatDate } from "@/lib/utils";
import { toast } from "sonner";

function useDeleteGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => goalService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.goals.all() });
      toast.success("Goal deleted");
    },
  });
}

const STATUS_CONFIG = {
  active:    { icon: TrendingUp,   color: "text-blue-600",   bg: "bg-blue-100",   bar: "bg-blue-500",   label: "Active"    },
  completed: { icon: CheckCircle,  color: "text-green-600",  bg: "bg-green-100",  bar: "bg-green-500",  label: "Completed" },
  paused:    { icon: PauseCircle,  color: "text-yellow-600", bg: "bg-yellow-100", bar: "bg-yellow-400", label: "Paused"    },
};

export default function GoalsPage() {
  const [statusFilter, setStatusFilter]       = useState<string | undefined>(undefined);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [contributeGoalId, setContributeGoalId] = useState<string | null>(null);
  const [contributeAmount, setContributeAmount] = useState("");

  const { data: goals = [], isLoading } = useGoals(statusFilter);
  const { mutate: deleteGoal, isPending: isDeleting } = useDeleteGoal();

  // We need a generic contribute hook — build it inline
  const qc = useQueryClient();
  const { mutate: contribute, isPending: isContributing } = useMutation({
    mutationFn: ({ id, amount }: { id: string; amount: number }) =>
      goalService.contribute(id, amount),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.goals.all() });
      toast.success("Contribution added!");
      setContributeGoalId(null);
      setContributeAmount("");
    },
  });

  const filters = ["All", "active", "completed", "paused"] as const;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Goals</h1>
          <p className="text-sm text-muted-foreground">
            {goals.length} goal{goals.length !== 1 ? "s" : ""}
          </p>
        </div>
        <Link
          href="/goals/new"
          className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="h-4 w-4" />
          New Goal
        </Link>
      </div>

      {/* Status filter */}
      <div className="flex items-center gap-2">
        {filters.map((f) => (
          <button
            key={f}
            onClick={() => setStatusFilter(f === "All" ? undefined : f)}
            className={`rounded-lg px-3 py-1.5 text-sm font-medium capitalize transition-colors ${
              (f === "All" && !statusFilter) || f === statusFilter
                ? "bg-primary text-primary-foreground"
                : "border border-border hover:bg-muted"
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      {/* Goal cards */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : goals.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-border bg-card py-20 shadow-sm text-center">
          <Target className="h-10 w-10 text-muted-foreground mb-3" />
          <p className="text-sm font-medium text-muted-foreground">No goals yet</p>
          <Link href="/goals/new" className="mt-3 text-sm font-medium text-primary hover:underline">
            Create your first goal →
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {goals.map((goal) => {
            const status = goal.status as keyof typeof STATUS_CONFIG;
            const cfg    = STATUS_CONFIG[status] ?? STATUS_CONFIG.active;
            const Icon   = cfg.icon;
            const pct    = Math.min(goal.progress_percent ?? 0, 100);

            return (
              <div key={goal.id} className="rounded-xl border border-border bg-card p-5 shadow-sm space-y-4">
                {/* Top row */}
                <div className="flex items-start justify-between gap-2">
                  <div className="min-w-0">
                    <p className="font-semibold truncate">{goal.name}</p>
                    {goal.description && (
                      <p className="text-xs text-muted-foreground mt-0.5 truncate">{goal.description}</p>
                    )}
                  </div>
                  <div className="flex items-center gap-1.5 flex-shrink-0">
                    <span className={`flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${cfg.bg} ${cfg.color}`}>
                      <Icon className="h-3 w-3" />
                      {cfg.label}
                    </span>
                    <button
                      onClick={() => setConfirmDeleteId(goal.id)}
                      className="rounded p-1 hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </button>
                  </div>
                </div>

                {/* Progress */}
                <div>
                  <div className="mb-1.5 flex items-center justify-between text-xs">
                    <span className="text-muted-foreground">
                      {formatCurrency(goal.current_amount, goal.currency)}
                      <span className="text-muted-foreground/60"> / {formatCurrency(goal.target_amount, goal.currency)}</span>
                    </span>
                    <span className="font-semibold">{pct.toFixed(0)}%</span>
                  </div>
                  <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                    <div
                      className={`h-full rounded-full transition-all duration-500 ${cfg.bar}`}
                      style={{ width: `${pct}%` }}
                    />
                  </div>
                </div>

                {/* Meta row */}
                <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Calendar className="h-3.5 w-3.5" />
                    <span>{formatDate(new Date(goal.target_date))}</span>
                  </div>
                  <div className="flex items-center gap-1 justify-end">
                    <TrendingUp className="h-3.5 w-3.5" />
                    <span>{formatCurrency(goal.monthly_savings_needed, goal.currency)}/mo needed</span>
                  </div>
                </div>

                {/* Days remaining */}
                {goal.days_remaining > 0 && (
                  <p className="text-xs text-muted-foreground border-t border-border pt-3">
                    <span className="font-medium text-foreground">{goal.days_remaining}</span> days remaining
                  </p>
                )}

                {/* Contribute button */}
                {goal.status === "active" && (
                  <button
                    onClick={() => { setContributeGoalId(goal.id); setContributeAmount(""); }}
                    className="w-full rounded-lg border border-primary text-primary px-4 py-1.5 text-sm font-medium hover:bg-primary/5 transition-colors"
                  >
                    + Add Contribution
                  </button>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Contribute modal */}
      {contributeGoalId && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
          onClick={() => setContributeGoalId(null)}
        >
          <div
            className="mx-4 w-full max-w-sm rounded-xl border border-border bg-background p-6 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 className="text-base font-semibold">Add Contribution</h2>
            <p className="mt-1 text-sm text-muted-foreground">
              How much are you adding to this goal?
            </p>
            <input
              type="number"
              min="0.01"
              step="0.01"
              placeholder="0.00"
              value={contributeAmount}
              onChange={(e) => setContributeAmount(e.target.value)}
              className="mt-4 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            />
            <div className="mt-4 flex gap-3">
              <button
                onClick={() => setContributeGoalId(null)}
                className="flex-1 rounded-lg border border-border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
              >
                Cancel
              </button>
              <button
                disabled={isContributing || !contributeAmount || Number(contributeAmount) <= 0}
                onClick={() => contribute({ id: contributeGoalId, amount: Number(contributeAmount) })}
                className="flex flex-1 items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-60 transition-colors"
              >
                {isContributing && <Loader2 className="h-4 w-4 animate-spin" />}
                Confirm
              </button>
            </div>
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
            <h2 className="text-base font-semibold">Delete goal?</h2>
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
                onClick={() => deleteGoal(confirmDeleteId, { onSuccess: () => setConfirmDeleteId(null) })}
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
