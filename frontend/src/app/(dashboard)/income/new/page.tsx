"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { ArrowLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { useCreateIncome } from "@/hooks/useExpenses";
import type { CreateIncomeRequest } from "@/types";

const CATEGORIES = [
  "Salary",
  "Freelance",
  "Business",
  "Investment",
  "Rental",
  "Interest",
  "Dividend",
  "Gift",
  "Other",
];

const PAYMENT_METHODS = [
  { label: "Bank Transfer", value: "bank" },
  { label: "Cash", value: "cash" },
  { label: "UPI", value: "upi" },
  { label: "Cheque", value: "cheque" },
  { label: "Wallet", value: "wallet" },
  { label: "Other", value: "other" },
];

export default function NewIncomePage() {
  const router = useRouter();
  const { mutate: createIncome, isPending } = useCreateIncome();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateIncomeRequest>({
    defaultValues: {
      date: new Date().toISOString().split("T")[0],
      currency: "INR",
      is_taxable: false,
    },
  });

  const onSubmit = (data: CreateIncomeRequest) => {
    createIncome(
      {
        ...data,
        amount: Number(data.amount),
        date: new Date(data.date + "T00:00:00").toISOString(),
      },
      { onSuccess: () => router.push("/income") }
    );
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Back + title */}
      <div className="flex items-center gap-3">
        <Link
          href="/income"
          className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h1 className="text-xl font-bold">Add Income</h1>
          <p className="text-xs text-muted-foreground">
            Record a new income entry
          </p>
        </div>
      </div>

      {/* Form */}
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="rounded-xl border border-border bg-card p-6 space-y-5 shadow-sm"
      >
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
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary uppercase"
              {...register("currency")}
            />
          </div>
        </div>

        {/* Source */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">
            Source <span className="text-destructive">*</span>
          </label>
          <input
            type="text"
            placeholder="e.g. Company name, Client name"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("source", { required: "Source is required" })}
          />
          {errors.source && (
            <p className="text-xs text-destructive">{errors.source.message}</p>
          )}
        </div>

        {/* Category */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Category</label>
          <select
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("category")}
          >
            <option value="">Select a category</option>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
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

        {/* Date + Payment method */}
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
                <option key={m.value} value={m.value}>{m.label}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Taxable */}
        <div className="flex items-center gap-3">
          <input
            type="checkbox"
            id="is_taxable"
            className="h-4 w-4 accent-primary"
            {...register("is_taxable")}
          />
          <label htmlFor="is_taxable" className="text-sm font-medium cursor-pointer">
            This income is taxable
          </label>
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
            href="/income"
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
            Save Income
          </button>
        </div>
      </form>
    </div>
  );
}
