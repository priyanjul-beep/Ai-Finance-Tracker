"use client";

import { useState, useCallback } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { ArrowLeft, Loader2, Mic, ScanLine } from "lucide-react";
import Link from "next/link";
import { useCreateExpense } from "@/hooks/useExpenses";
import { VoiceExpenseModal } from "@/components/expenses/VoiceExpenseModal";
import { ReceiptScanModal } from "@/components/expenses/ReceiptScanModal";
import type { CreateExpenseRequest, AIVoiceParseResponse, AIReceiptScanResponse } from "@/types";

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

/** Map Gemini category tokens → form select options */
const AI_CATEGORY_MAP: Record<string, string> = {
  food:          "Food & Dining",
  travel:        "Travel",
  shopping:      "Shopping",
  entertainment: "Entertainment",
  health:        "Healthcare",
  investment:    "Other",
  education:     "Education",
  bills:         "Utilities",
  recharge:      "Utilities",
  fuel:          "Transportation",
  rent:          "Housing",
  salary:        "Other",
  utilities:     "Utilities",
  subscription:  "Subscriptions",
  personal_care: "Personal Care",
  gift:          "Other",
  charitable:    "Other",
  insurance:     "Other",
  others:        "Other",
  unknown:       "Other",
  transportation:"Transportation",
  transport:     "Transportation",
};

/** Resolve "today" / "yesterday" relative date strings → YYYY-MM-DD */
function resolveDateString(raw: string): string {
  if (!raw) return new Date().toISOString().split("T")[0];
  const lower = raw.toLowerCase().trim();
  if (lower === "today")     return new Date().toISOString().split("T")[0];
  if (lower === "yesterday") {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    return d.toISOString().split("T")[0];
  }
  // Try to parse ISO or other formats
  try {
    const d = new Date(raw);
    if (!isNaN(d.getTime())) return d.toISOString().split("T")[0];
  } catch {}
  return new Date().toISOString().split("T")[0];
}

export default function NewExpensePage() {
  const router = useRouter();
  const { mutate: createExpense, isPending } = useCreateExpense();
  const [voiceOpen, setVoiceOpen] = useState(false);
  const [scanOpen, setScanOpen]   = useState(false);
  const [autoFillBanner, setAutoFillBanner] = useState(false);

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<CreateExpenseRequest>({
    defaultValues: {
      date: new Date().toISOString().split("T")[0],
      currency: "INR",
      expense_type: "spend",
    },
  });

  /** Called by VoiceExpenseModal when parsing succeeds — auto-fills every field */
  const handleAutoFill = useCallback(
    (result: AIVoiceParseResponse) => {
      if (result.amount)         setValue("amount",         result.amount);
      if (result.merchant)       setValue("merchant",       result.merchant);
      if (result.notes)          setValue("notes",          result.notes);
      if (result.expense_type)   setValue("expense_type",   result.expense_type as CreateExpenseRequest["expense_type"]);
      if (result.payment_method) setValue("payment_method", result.payment_method);

      // Resolve AI category token → human-readable form option
      const cat = AI_CATEGORY_MAP[result.category?.toLowerCase?.() ?? ""] ?? "Other";
      setValue("category", cat);

      // Resolve relative date → YYYY-MM-DD
      setValue("date", resolveDateString(result.date));

      setAutoFillBanner(true);
      setTimeout(() => setAutoFillBanner(false), 5_000);
    },
    [setValue]
  );

  /** Called by ReceiptScanModal — maps receipt scan result to the form fields */
  const handleScanFill = useCallback(
    (result: AIReceiptScanResponse) => {
      if (result.amount)         setValue("amount",         result.amount);
      if (result.merchant)       setValue("merchant",       result.merchant);
      if (result.notes)          setValue("notes",          result.notes);
      if (result.expense_type)   setValue("expense_type",   result.expense_type as CreateExpenseRequest["expense_type"]);
      if (result.payment_method) setValue("payment_method", result.payment_method);
      if (result.currency)       setValue("currency",       result.currency);

      const cat = AI_CATEGORY_MAP[result.category?.toLowerCase?.() ?? ""] ?? "Other";
      setValue("category", cat);
      setValue("date", resolveDateString(result.date));

      setAutoFillBanner(true);
      setTimeout(() => setAutoFillBanner(false), 5_000);
    },
    [setValue]
  );

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
        <div className="flex-1">
          <h1 className="text-xl font-bold">Add Expense</h1>
          <p className="text-xs text-muted-foreground">
            Fill in the details below to record a new expense
          </p>
        </div>

        {/* Voice button */}
        <button
          type="button"
          onClick={() => setVoiceOpen(true)}
          className="flex items-center gap-2 rounded-xl border border-primary/30 bg-primary/5 px-3 py-2 text-sm font-medium text-primary hover:bg-primary/10 transition-colors"
        >
          <Mic className="h-4 w-4" />
          <span className="hidden sm:inline">Voice</span>
        </button>

        {/* Scan receipt button */}
        <button
          type="button"
          onClick={() => setScanOpen(true)}
          className="flex items-center gap-2 rounded-xl border border-violet-500/30 bg-violet-500/5 px-3 py-2 text-sm font-medium text-violet-600 dark:text-violet-400 hover:bg-violet-500/10 transition-colors"
        >
          <ScanLine className="h-4 w-4" />
          <span className="hidden sm:inline">Scan</span>
        </button>
      </div>

      {/* Auto-fill success banner */}
      {autoFillBanner && (
        <div className="flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-2.5 text-sm text-emerald-700 dark:border-emerald-800/50 dark:bg-emerald-950/30 dark:text-emerald-400">
          <span className="h-2 w-2 rounded-full bg-emerald-500 animate-pulse" />
          Form auto-filled from AI — review and save.
        </div>
      )}

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
              placeholder="INR"
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

      {/* Voice expense modal */}
      <VoiceExpenseModal
        open={voiceOpen}
        onClose={() => setVoiceOpen(false)}
        onAutoFill={handleAutoFill}
      />

      {/* Receipt scan modal */}
      <ReceiptScanModal
        open={scanOpen}
        onClose={() => setScanOpen(false)}
        onAutoFill={handleScanFill}
      />
    </div>
  );
}
