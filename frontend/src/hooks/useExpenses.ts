"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import {
  expenseService,
  incomeService,
  budgetService,
  subscriptionService,
  goalService,
} from "@/services/expense.service";
import { queryKeys } from "@/lib/query-client";
import type {
  CreateExpenseRequest,
  UpdateExpenseRequest,
  ExpenseFilters,
  CreateIncomeRequest,
  CreateBudgetRequest,
  CreateSubscriptionRequest,
  CreateGoalRequest,
} from "@/types";

// ─── Expenses ─────────────────────────────────────────────────────────────────

export function useExpenses(filters: ExpenseFilters = {}) {
  return useQuery({
    queryKey: queryKeys.expenses.list(filters as Record<string, unknown>),
    queryFn: () => expenseService.list(filters),
    staleTime: 30_000,
  });
}

export function useExpense(id: string) {
  return useQuery({
    queryKey: queryKeys.expenses.detail(id),
    queryFn: () => expenseService.getById(id),
    enabled: !!id,
  });
}

export function useCreateExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateExpenseRequest) => expenseService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.expenses.all() });
      qc.invalidateQueries({ queryKey: queryKeys.analytics.dashboard() });
      toast.success("Expense added");
    },
  });
}

export function useUpdateExpense(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateExpenseRequest) =>
      expenseService.update(id, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.expenses.all() });
      qc.invalidateQueries({ queryKey: queryKeys.analytics.dashboard() });
      toast.success("Expense updated");
    },
  });
}

export function useDeleteExpense() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => expenseService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.expenses.all() });
      qc.invalidateQueries({ queryKey: queryKeys.analytics.dashboard() });
      toast.success("Expense deleted");
    },
  });
}

export function useSearchExpenses(query: string) {
  return useQuery({
    queryKey: ["expenses", "search", query],
    queryFn: () => expenseService.search(query),
    enabled: query.length > 2,
    staleTime: 10_000,
  });
}

// ─── Income ──────────────────────────────────────────────────────────────────

export function useIncomeList(filters: Record<string, unknown> = {}) {
  return useQuery({
    queryKey: queryKeys.income.list(filters),
    queryFn: () => incomeService.list(filters),
    staleTime: 30_000,
  });
}

export function useCreateIncome() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateIncomeRequest) => incomeService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.income.all() });
      qc.invalidateQueries({ queryKey: queryKeys.analytics.dashboard() });
      toast.success("Income added");
    },
  });
}

export function useDeleteIncome() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => incomeService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.income.all() });
      qc.invalidateQueries({ queryKey: queryKeys.analytics.dashboard() });
      toast.success("Income deleted");
    },
  });
}

// ─── Budgets ─────────────────────────────────────────────────────────────────

export function useBudgets(year?: number, month?: number, category?: string) {
  return useQuery({
    queryKey: queryKeys.budgets.list(year, month, category),
    queryFn: () => budgetService.list(year, month, category),
    staleTime: 60_000,
  });
}

export function useCreateBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateBudgetRequest) => budgetService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.budgets.all() });
      toast.success("Budget created");
    },
  });
}

export function useDeleteBudget() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => budgetService.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.budgets.all() });
      toast.success("Budget deleted");
    },
  });
}

// ─── Subscriptions ───────────────────────────────────────────────────────────

export function useSubscriptions(activeOnly = false) {
  return useQuery({
    queryKey: queryKeys.subscriptions.list(activeOnly),
    queryFn: () => subscriptionService.list(activeOnly),
    staleTime: 60_000,
  });
}

export function useUpcomingSubscriptions(days = 7) {
  return useQuery({
    queryKey: queryKeys.subscriptions.upcoming(days),
    queryFn: () => subscriptionService.getUpcoming(days),
    staleTime: 60_000,
  });
}

export function useCreateSubscription() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateSubscriptionRequest) =>
      subscriptionService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.subscriptions.all() });
      toast.success("Subscription added");
    },
  });
}

// ─── Goals ───────────────────────────────────────────────────────────────────

export function useGoals(status?: string) {
  return useQuery({
    queryKey: queryKeys.goals.list(status),
    queryFn: () => goalService.list(status),
    staleTime: 60_000,
  });
}

export function useCreateGoal() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateGoalRequest) => goalService.create(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.goals.all() });
      toast.success("Goal created");
    },
  });
}

export function useContributeToGoal(id: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (amount: number) => goalService.contribute(id, amount),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.goals.all() });
      toast.success("Contribution added!");
    },
  });
}
