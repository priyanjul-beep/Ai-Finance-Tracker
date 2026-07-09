"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { ArrowLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { useCreateExpense } from "@/hooks/useExpenses";
import type { CreateExpenseRequest } from "@/types";

const CATEGORIES = [
  "Food & Dining",
  "Transportation",
  "Shopping",
  "Entertainment",
  "Healthcare",
  "Housing",
  "Utilities",
  "Travel",
  "Education",
  "Personal Care",
  "Subscriptions",
  "Other",
];

const PAYMENT_METHODS: { label: string; value: string }[] = [
  { label: "Cash", value: "cash" },
  { label: "Credit / Debit Card", value: "card" },
  { label: "UPI", value: "upi" },
  { label: "Bank Transfer", value: "bank" },
  { label: "Wallet", value: "wallet" },
  { label: "Online", value: "online" },
];

export default function NewExpensePage() {
  const router = useRouter();
  const { mutate: createExpense, isPending } = useCreateExpense();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateExpenseRequest>({
    defaultValues: {
      date: new Date().toISOString().split("T")[0],
      currency: "USD",
      expense_type: "spend",
    },
  });

  const onSubmit = (data: CreateExpenseRequest) => {
    createExpense(
      {
        ...data,
        amount: Number(data.amount),
        // Backend time.Time requires RFC3339; HTML date input gives YYYY-MM-DD
        date: new Date(data.date + "T00:00:00").toISOString(),
      },
      { onSuccess: () => router.push("/expenses") }
    );
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Back + title */}
      <div className="flex items-center gap-3">
        <Link
          href="/expenses"
          className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h1 className="text-xl font-bold">Add Expense</h1>
          <p className="text-xs text-muted-foreground">
            Fill in the details below to record a new expense
          </p>
        </div>
      </div>

      {/* Form card */}
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="rounded-xl border border-border bg-card p-6 space-y-5 shadow-sm"
      >
        {/* Amount + Currency row */}
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
              placeholder="USD"
              maxLength={3}
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary uppercase"
              {...register("currency")}
            />
          </div>
        </div>

        {/* Category */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">
            Category <span className="text-destructive">*</span>
          </label>
          <select
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("category", { required: "Category is required" })}
          >
            <option value="">Select a category</option>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {c}
              </option>
            ))}
          </select>
          {errors.category && (
            <p className="text-xs text-destructive">{errors.category.message}</p>
          )}
        </div>

        {/* Merchant */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Merchant</label>
          <input
            type="text"
            placeholder="e.g. Starbucks"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("merchant")}
          />
        </div>

        {/* Description */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Description</label>
          <input
            type="text"
            placeholder="Brief description"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("description")}
          />
        </div>

        {/* Date + Payment method row */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium">
              Date <span className="text-destructive">*</span>
            </label>
            <input
              type="date"
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("date", { required: "Date is required" })}
            />
            {errors.date && (
              <p className="text-xs text-destructive">{errors.date.message}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <label className="text-sm font-medium">Payment Method</label>
            <select
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("payment_method")}
            >
              <option value="">Select method</option>
              {PAYMENT_METHODS.map((m) => (
                <option key={m.value} value={m.value}>
                  {m.label}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Expense type */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Type</label>
          <div className="flex gap-3">
            {(["spend", "refund", "transfer"] as const).map((t) => (
              <label
                key={t}
                className="flex items-center gap-2 text-sm capitalize cursor-pointer"
              >
                <input
                  type="radio"
                  value={t}
                  className="accent-primary"
                  {...register("expense_type")}
                />
                {t}
              </label>
            ))}
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
            href="/expenses"
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
            Save Expense
          </button>
        </div>
      </form>
    </div>
  );
}
