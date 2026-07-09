"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { ArrowLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { useCreateSubscription } from "@/hooks/useExpenses";
import type { CreateSubscriptionRequest } from "@/types";

const CATEGORIES = [
  "Entertainment", "Utilities", "Software", "Health",
  "Education", "Food", "Shopping", "Finance", "Other",
];

const BILLING_CYCLES = [
  { label: "Daily",     value: "daily"     },
  { label: "Weekly",    value: "weekly"    },
  { label: "Monthly",   value: "monthly"   },
  { label: "Quarterly", value: "quarterly" },
  { label: "Yearly",    value: "yearly"    },
];

const PAYMENT_METHODS = [
  { label: "Credit / Debit Card", value: "card"   },
  { label: "UPI",                 value: "upi"    },
  { label: "Bank Transfer",       value: "bank"   },
  { label: "Wallet",              value: "wallet" },
  { label: "Cash",                value: "cash"   },
];

export default function NewSubscriptionPage() {
  const router = useRouter();
  const { mutate: createSubscription, isPending } = useCreateSubscription();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateSubscriptionRequest>({
    defaultValues: {
      currency: "INR",
      billing_cycle: "monthly",
      next_billing_date: new Date().toISOString().split("T")[0],
    },
  });

  const onSubmit = (data: CreateSubscriptionRequest) => {
    createSubscription(
      {
        ...data,
        amount: Number(data.amount),
        next_billing_date: new Date(data.next_billing_date + "T00:00:00").toISOString(),
      },
      { onSuccess: () => router.push("/subscriptions") }
    );
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Back + title */}
      <div className="flex items-center gap-3">
        <Link
          href="/subscriptions"
          className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h1 className="text-xl font-bold">Add Subscription</h1>
          <p className="text-xs text-muted-foreground">Track a recurring payment</p>
        </div>
      </div>

      {/* Form */}
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="rounded-xl border border-border bg-card p-6 space-y-5 shadow-sm"
      >
        {/* Name */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">
            Name <span className="text-destructive">*</span>
          </label>
          <input
            type="text"
            placeholder="e.g. Netflix, Spotify"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("name", { required: "Name is required" })}
          />
          {errors.name && (
            <p className="text-xs text-destructive">{errors.name.message}</p>
          )}
        </div>

        {/* Amount + Currency */}
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-2 space-y-1.5">
            <label className="text-sm font-medium">
              Amount <span className="text-destructive">*</span>
            </label>
            <input
              type="number"
              step="0.01"
              min="0"
              placeholder="0.00"
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("amount", {
                required: "Amount is required",
                min: { value: 0.01, message: "Must be greater than 0" },
              })}
            />
            {errors.amount && (
              <p className="text-xs text-destructive">{errors.amount.message}</p>
            )}
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Currency</label>
            <input
              type="text"
              placeholder="INR"
              maxLength={3}
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm uppercase outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("currency")}
            />
          </div>
        </div>

        {/* Billing cycle + Next billing date */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Billing Cycle</label>
            <select
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("billing_cycle")}
            >
              {BILLING_CYCLES.map((c) => (
                <option key={c.value} value={c.value}>{c.label}</option>
              ))}
            </select>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium">
              Next Billing Date <span className="text-destructive">*</span>
            </label>
            <input
              type="date"
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("next_billing_date", { required: "Date is required" })}
            />
            {errors.next_billing_date && (
              <p className="text-xs text-destructive">{errors.next_billing_date.message}</p>
            )}
          </div>
        </div>

        {/* Category + Payment method */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Category</label>
            <select
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("category")}
            >
              <option value="">Select category</option>
              {CATEGORIES.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium">Payment Method</label>
            <select
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("payment_method")}
            >
              <option value="">Select method</option>
              {PAYMENT_METHODS.map((m) => (
                <option key={m.value} value={m.value}>{m.label}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Notes */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Notes</label>
          <textarea
            rows={3}
            placeholder="Any additional notes…"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary resize-none"
            {...register("notes")}
          />
        </div>

        {/* Actions */}
        <div className="flex gap-3 pt-2">
          <Link
            href="/subscriptions"
            className="flex-1 rounded-lg border border-border px-4 py-2 text-center text-sm font-medium hover:bg-muted transition-colors"
          >
            Cancel
          </Link>
          <button
            type="submit"
            disabled={isPending}
            className="flex flex-1 items-center justify-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60 transition-colors"
          >
            {isPending && <Loader2 className="h-4 w-4 animate-spin" />}
            Save Subscription
          </button>
        </div>
      </form>
    </div>
  );
}
