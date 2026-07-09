"use client";

import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { ArrowLeft, Loader2 } from "lucide-react";
import Link from "next/link";
import { useCreateGoal } from "@/hooks/useExpenses";
import type { CreateGoalRequest } from "@/types";

const CATEGORIES = [
  "Emergency Fund", "Vacation", "Home", "Car", "Education",
  "Retirement", "Wedding", "Business", "Health", "Gadgets", "Other",
];

export default function NewGoalPage() {
  const router = useRouter();
  const { mutate: createGoal, isPending } = useCreateGoal();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateGoalRequest>({
    defaultValues: {
      currency: "INR",
      priority: 3,
      current_amount: 0,
      target_date: new Date(new Date().setFullYear(new Date().getFullYear() + 1))
        .toISOString()
        .split("T")[0],
    },
  });

  const onSubmit = (data: CreateGoalRequest) => {
    createGoal(
      {
        ...data,
        target_amount:  Number(data.target_amount),
        current_amount: Number(data.current_amount ?? 0),
        priority:       Number(data.priority),
        target_date: new Date(data.target_date + "T00:00:00").toISOString(),
      },
      { onSuccess: () => router.push("/goals") }
    );
  };

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Back + title */}
      <div className="flex items-center gap-3">
        <Link
          href="/goals"
          className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h1 className="text-xl font-bold">Create Goal</h1>
          <p className="text-xs text-muted-foreground">Set a financial savings target</p>
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
            Goal Name <span className="text-destructive">*</span>
          </label>
          <input
            type="text"
            placeholder="e.g. Buy a car, Emergency fund"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("name", { required: "Name is required" })}
          />
          {errors.name && (
            <p className="text-xs text-destructive">{errors.name.message}</p>
          )}
        </div>

        {/* Description */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Description</label>
          <input
            type="text"
            placeholder="Optional description"
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
            {...register("description")}
          />
        </div>

        {/* Target + Current amount + Currency */}
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-1 space-y-1.5">
            <label className="text-sm font-medium">
              Target Amount <span className="text-destructive">*</span>
            </label>
            <input
              type="number"
              step="0.01"
              min="0.01"
              placeholder="0.00"
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("target_amount", {
                required: "Target amount is required",
                min: { value: 0.01, message: "Must be greater than 0" },
              })}
            />
            {errors.target_amount && (
              <p className="text-xs text-destructive">{errors.target_amount.message}</p>
            )}
          </div>

          <div className="col-span-1 space-y-1.5">
            <label className="text-sm font-medium">Already Saved</label>
            <input
              type="number"
              step="0.01"
              min="0"
              placeholder="0.00"
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("current_amount")}
            />
          </div>

          <div className="col-span-1 space-y-1.5">
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

        {/* Target date + Category */}
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-1.5">
            <label className="text-sm font-medium">
              Target Date <span className="text-destructive">*</span>
            </label>
            <input
              type="date"
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
              {...register("target_date", { required: "Target date is required" })}
            />
            {errors.target_date && (
              <p className="text-xs text-destructive">{errors.target_date.message}</p>
            )}
          </div>

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
        </div>

        {/* Priority */}
        <div className="space-y-1.5">
          <label className="text-sm font-medium">Priority (1 = Highest, 5 = Lowest)</label>
          <div className="flex gap-3">
            {[1, 2, 3, 4, 5].map((p) => (
              <label key={p} className="flex items-center gap-1.5 text-sm cursor-pointer">
                <input
                  type="radio"
                  value={p}
                  className="accent-primary"
                  {...register("priority")}
                />
                {p}
              </label>
            ))}
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-3 pt-2">
          <Link
            href="/goals"
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
            Create Goal
          </button>
        </div>
      </form>
    </div>
  );
}
