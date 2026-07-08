import { QueryClient } from "@tanstack/react-query";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60 * 1000,          // 1 minute
      gcTime: 5 * 60 * 1000,         // 5 minutes (formerly cacheTime)
      retry: (failureCount, error) => {
        // Don't retry on 4xx errors
        if (error instanceof Error && "status" in error) {
          const status = (error as { status: number }).status;
          if (status >= 400 && status < 500) return false;
        }
        return failureCount < 2;
      },
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
});

// Query key factories – strongly typed, co-located
export const queryKeys = {
  // Auth
  profile: () => ["profile"] as const,

  // Expenses
  expenses: {
    all: () => ["expenses"] as const,
    list: (params?: Record<string, unknown>) =>
      ["expenses", "list", params] as const,
    detail: (id: string) => ["expenses", "detail", id] as const,
    duplicates: (id: string) => ["expenses", "duplicates", id] as const,
  },

  // Income
  income: {
    all: () => ["income"] as const,
    list: (params?: Record<string, unknown>) =>
      ["income", "list", params] as const,
    detail: (id: string) => ["income", "detail", id] as const,
    monthlyTotal: (year: number, month: number) =>
      ["income", "monthly", year, month] as const,
  },

  // Budgets
  budgets: {
    all: () => ["budgets"] as const,
    list: (year?: number, month?: number) =>
      ["budgets", "list", year, month] as const,
    detail: (id: string) => ["budgets", "detail", id] as const,
  },

  // Subscriptions
  subscriptions: {
    all: () => ["subscriptions"] as const,
    list: (activeOnly?: boolean) =>
      ["subscriptions", "list", activeOnly] as const,
    detail: (id: string) => ["subscriptions", "detail", id] as const,
    upcoming: (days?: number) => ["subscriptions", "upcoming", days] as const,
  },

  // Goals
  goals: {
    all: () => ["goals"] as const,
    list: (status?: string) => ["goals", "list", status] as const,
    detail: (id: string) => ["goals", "detail", id] as const,
  },

  // Analytics
  analytics: {
    dashboard: () => ["analytics", "dashboard"] as const,
    monthly: (year: number, month: number) =>
      ["analytics", "monthly", year, month] as const,
    yearly: (year: number) => ["analytics", "yearly", year] as const,
    predictions: () => ["analytics", "predictions"] as const,
    insights: () => ["analytics", "insights"] as const,
    healthScore: () => ["analytics", "health-score"] as const,
  },

  // Tags
  tags: {
    all: () => ["tags"] as const,
  },

  // Notifications
  notifications: {
    all: () => ["notifications"] as const,
    list: (page?: number) => ["notifications", "list", page] as const,
  },
};
