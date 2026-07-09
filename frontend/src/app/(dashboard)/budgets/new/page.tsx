"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { ArrowLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { useCreateBudget } from "@/hooks/useExpenses";
import type { CreateBudgetRequest } from "@/types";

const CATEGORIES = [
  "Food & Dining","Transportation","Shopping","Entertainment",
  "Healthcare","Housing","Utilities","Travel","Education",
  "Personal Care","Subscriptions","Other",
];

const MONTHS = [
  "January","February","March","April","May","June",
  "July","August","September","October","November","December",
];

const currentYear  = new Date().getFullYear();
const currentMonth = new Date().getMonth() + 1;

export default function NewBudgetPage() {
  const router = useRouter();
  const { mutate: createBudget, isPending } = useCreateBudget();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateBudgetRequest>({
    defaultValues: {
      currency: "INR",
      month: currentMonth,
      year:  currentYear,
      period: "monthly",
      alert_at: 80,
    },
  });

  const onSubmit = (data: CreateBudgetRequest) => {
    createBudget(
      { ...data, amount: Number(data.amount), month: Number(data.month), year: Number(data.year) },
      { onSuccess: () => router.push("/budgets") }
    );
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Back + title */}
      <div className="flex items-center gap-3">
        <Link
          href="/budgets"
          className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h1 className="text-xl font-bold">Create Budget</h1>
          <p className="text-xs text-muted-foreground">Set a spending limit for a category</p>
        </div>
      </div>

      {/* Form */}
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="rounded-xl border border-border bg-card p-6 space-y-5 shadow-sm"
      >
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
              <option key={c} value={c}>{c}</option>
            ))}
          </select>
          {errors.category && (
            <p className="text-xs text-destructive">{errors.category.message}</p>
          )}
        </div>

        {/* Amount + Currency */}
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-2 space-y-1.5">
            <label className="text-sm font-medium">
              Budget Amount <span className="text-destructive">*</span>
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

        {/* Month + Year */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium">Month</label>
            <select
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("month")}
            >
              {MONTHS.map((m, i) => (
                <option key={m} value={i + 1}>{m}</option>
              ))}
            </select>
          </div>
          <div className="space-y-1.5">
            <label className="text-sm font-medium">
              Year <span className="text-destructive">*</span>
            </label>
            <select
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("year", { required: true })}
            >
              {[currentYear - 1, currentYear, currentYear + 1].map((y) => (
                <option key={y} value={y}>{y}</option>
              ))}
            </select>
          </div>
        </div>

        {/* Alert threshold */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">
            Alert me when spending reaches (%)
          </label>
          <input
            type="number"
            min="1"
            max="100"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("alert_at", { min: 1, max: 100 })}
          />
          <p className="text-xs text-muted-foreground">
            You'll be warned when this % of the budget is consumed. Default is 80%.
          </p>
        </div>

        {/* Description */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Description</label>
          <input
            type="text"
            placeholder="Optional note about this budget"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("description")}
          />
        </div>

        {/* Actions */}
        <div className="flex gap-3 pt-2">
          <Link
            href="/budgets"
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
            Create Budget
          </button>
        </div>
      </form>
    </div>
  );
}
