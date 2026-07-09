"use client";

import { useState } from "react";
import Link from "next/link";
import { Plus, Loader2, Trash2, Calendar, RefreshCw } from "lucide-react";
import { useSubscriptions, useUpcomingSubscriptions } from "@/hooks/useExpenses";
import { subscriptionService } from "@/services/expense.service";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { queryKeys } from "@/lib/query-client";
import { formatCurrency, formatDate } from "@/lib/utils";
import { toast } from "sonner";

const CYCLE_LABELS: Record<string, string> = {
  daily: "Daily",
  weekly: "Weekly",
  monthly: "Monthly",
  quarterly: "Quarterly",
  yearly: "Yearly",
};

const CATEGORY_COLORS: Record<string, string> = {
  Entertainment: "bg-purple-100 text-purple-700",
  Utilities: "bg-cyan-100 text-cyan-700",
  Software: "bg-blue-100 text-blue-700",
  Health: "bg-green-100 text-green-700",
  Education: "bg-yellow-100 text-yellow-700",
  Food: "bg-orange-100 text-orange-700",
  Shopping: "bg-pink-100 text-pink-700",
  Finance: "bg-teal-100 text-teal-700",
  Other: "bg-gray-100 text-gray-700",
};

function useDeleteSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => subscriptionService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.subscriptions.all() });
      toast.success("Subscription deleted");
    },
  });
}

export default function SubscriptionsPage() {
  const [showActiveOnly, setShowActiveOnly] = useState(false);
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

  const { data: subscriptions = [], isLoading } = useSubscriptions(showActiveOnly);
  const { data: upcoming = [] } = useUpcomingSubscriptions(30);
  const { mutate: deleteSubscription, isPending: isDeleting } = useDeleteSubscription();

  const totalMonthly = subscriptions
    .filter((s) => s.is_active && s.billing_cycle === "monthly")
    .reduce((sum, s) => sum + s.amount, 0);

  const currency = subscriptions[0]?.currency ?? "INR";

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Subscriptions</h1>
          <p className="text-sm text-muted-foreground">
            {subscriptions.length} subscription{subscriptions.length !== 1 ? "s" : ""}
            {totalMonthly > 0 && ` · ${formatCurrency(totalMonthly, currency)}/mo`}
          </p>
        </div>
        <Link
          href="/subscriptions/new"
          className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="h-4 w-4" />
          Add Subscription
        </Link>
      </div>

      {/* Upcoming renewals */}
      {upcoming.length > 0 && (
        <div className="rounded-xl border border-border bg-card p-5 shadow-sm">
          <div className="flex items-center gap-2 mb-3">
            <Calendar className="h-4 w-4 text-primary" />
            <h2 className="text-sm font-semibold">Upcoming Renewals (next 30 days)</h2>
          </div>
          <div className="flex flex-wrap gap-2">
            {upcoming.map((s) => (
              <div
                key={s.id}
                className="flex items-center gap-2 rounded-lg border border-border bg-muted/40 px-3 py-2 text-sm"
              >
                <span className="font-medium">{s.name}</span>
                <span className="text-muted-foreground">·</span>
                <span className="text-destructive font-medium">
                  {formatCurrency(s.amount, s.currency)}
                </span>
                <span className="text-muted-foreground">·</span>
                <span className="text-xs text-muted-foreground">
                  {formatDate(new Date(s.next_billing_date))}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Filter toggle */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => setShowActiveOnly(false)}
          className={`rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
            !showActiveOnly
              ? "bg-primary text-primary-foreground"
              : "border border-border hover:bg-muted"
          }`}
        >
          All
        </button>
        <button
          onClick={() => setShowActiveOnly(true)}
          className={`rounded-lg px-3 py-1.5 text-sm font-medium transition-colors ${
            showActiveOnly
              ? "bg-primary text-primary-foreground"
              : "border border-border hover:bg-muted"
          }`}
        >
          Active only
        </button>
      </div>

      {/* List */}
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : subscriptions.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-border bg-card py-20 shadow-sm text-center">
          <p className="text-sm font-medium text-muted-foreground">No subscriptions yet</p>
          <Link href="/subscriptions/new" className="mt-3 text-sm font-medium text-primary hover:underline">
            Add your first subscription →
          </Link>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
          {subscriptions.map((sub) => (
            <div
              key={sub.id}
              className={`rounded-xl border bg-card p-5 shadow-sm space-y-3 transition-opacity ${
                sub.is_active ? "border-border" : "border-border opacity-60"
              }`}
            >
              {/* Top row */}
              <div className="flex items-start justify-between gap-2">
                <div className="min-w-0">
                  <p className="font-semibold truncate">{sub.name}</p>
                  <div className="mt-1 flex items-center gap-2 flex-wrap">
                    <span
                      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${
                        CATEGORY_COLORS[sub.category] ?? "bg-gray-100 text-gray-700"
                      }`}
                    >
                      {sub.category || "Other"}
                    </span>
                    <span
                      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
                        sub.is_active
                          ? "bg-green-100 text-green-700"
                          : "bg-gray-100 text-gray-500"
                      }`}
                    >
                      <span className={`h-1.5 w-1.5 rounded-full ${sub.is_active ? "bg-green-500" : "bg-gray-400"}`} />
                      {sub.is_active ? "Active" : "Inactive"}
                    </span>
                  </div>
                </div>
                <button
                  onClick={() => setConfirmDeleteId(sub.id)}
                  className="flex-shrink-0 rounded p-1.5 hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </button>
              </div>

              {/* Amount + cycle */}
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-1.5 text-sm text-muted-foreground">
                  <RefreshCw className="h-3.5 w-3.5" />
                  <span>{CYCLE_LABELS[sub.billing_cycle] ?? sub.billing_cycle}</span>
                </div>
                <span className="text-lg font-bold">
                  {formatCurrency(sub.amount, sub.currency)}
                </span>
              </div>

              {/* Next billing */}
              <div className="flex items-center gap-1.5 text-xs text-muted-foreground border-t border-border pt-3">
                <Calendar className="h-3.5 w-3.5" />
                <span>
                  Next billing: <span className="font-medium text-foreground">
                    {formatDate(new Date(sub.next_billing_date))}
                  </span>
                </span>
              </div>

              {sub.notes && (
                <p className="text-xs text-muted-foreground truncate">{sub.notes}</p>
              )}
            </div>
          ))}
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
            <h2 className="text-base font-semibold">Delete subscription?</h2>
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
                  deleteSubscription(confirmDeleteId, {
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
