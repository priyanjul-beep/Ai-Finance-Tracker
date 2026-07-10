import api from "./api";
import { buildQueryString } from "@/lib/utils";
import type {
  Expense,
  CreateExpenseRequest,
  UpdateExpenseRequest,
  AIExpenseParsed,
  PaginatedResponse,
  ExpenseFilters,
} from "@/types";

export const expenseService = {
  // Create an expense
  create: async (data: CreateExpenseRequest): Promise<Expense> => {
    const res = await api.post<Expense>("/expenses", data);
    return res.data;
  },

  // Get a single expense
  getById: async (id: string): Promise<Expense> => {
    const res = await api.get<Expense>(`/expenses/${id}`);
    return res.data;
  },

  // List expenses with filters + pagination
  list: async (
    filters: ExpenseFilters = {}
  ): Promise<PaginatedResponse<Expense>> => {
    const res = await api.get<{
      data: Expense[];
      pagination?: {
        page: number;
        limit: number;
        total: number;
        total_pages: number;
        has_next: boolean;
        has_prev: boolean;
      };
      // flat shape (legacy / other endpoints)
      total?: number;
      total_pages?: number;
      page?: number;
      limit?: number;
    }>(`/expenses${buildQueryString(filters)}`);

    const raw = res.data;
    const p = raw.pagination;
    return {
      data:        raw.data ?? [],
      total:       p?.total       ?? raw.total       ?? 0,
      total_pages: p?.total_pages ?? raw.total_pages ?? 1,
      page:        p?.page        ?? raw.page        ?? 1,
      limit:       p?.limit       ?? raw.limit       ?? 10,
    };
  },

  // Update an expense
  update: async (id: string, data: UpdateExpenseRequest): Promise<Expense> => {
    const res = await api.put<Expense>(`/expenses/${id}`, data);
    return res.data;
  },

  // Delete an expense
  delete: async (id: string): Promise<void> => {
    await api.delete(`/expenses/${id}`);
  },

  // AI-parse expense from text or image URL
  parseWithAI: async (
    text?: string,
    imageUrl?: string
  ): Promise<AIExpenseParsed> => {
    const res = await api.post<AIExpenseParsed>("/expenses/ai-parse", {
      text,
      image_url: imageUrl,
    });
    return res.data;
  },

  // Natural-language search
  search: async (query: string): Promise<Expense[]> => {
    const res = await api.get<{ expenses: Expense[] }>(
      `/expenses/search?q=${encodeURIComponent(query)}`
    );
    return res.data.expenses;
  },

  // Get duplicate candidates for an expense
  getDuplicates: async (id: string): Promise<Expense[]> => {
    const res = await api.get<{ expenses: Expense[] }>(
      `/expenses/${id}/duplicates`
    );
    return res.data.expenses;
  },

  // Upload receipt image
  uploadReceipt: async (id: string, file: File): Promise<Expense> => {
    const form = new FormData();
    form.append("receipt", file);
    const res = await api.post<Expense>(`/expenses/${id}/receipt`, form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return res.data;
  },
};

export const incomeService = {
  create: async (data: import("@/types").CreateIncomeRequest) => {
    const res = await api.post<import("@/types").Income>("/income", data);
    return res.data;
  },
  getById: async (id: string) => {
    const res = await api.get<import("@/types").Income>(`/income/${id}`);
    return res.data;
  },
  list: async (filters: import("@/types").IncomeFilters = {}) => {
    const res = await api.get<PaginatedResponse<import("@/types").Income>>(
      `/income${buildQueryString(filters)}`
    );
    return res.data;
  },
  update: async (id: string, data: import("@/types").UpdateIncomeRequest) => {
    const res = await api.put<import("@/types").Income>(`/income/${id}`, data);
    return res.data;
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/income/${id}`);
  },
};

export const budgetService = {
  create: async (data: import("@/types").CreateBudgetRequest) => {
    const res = await api.post<import("@/types").BudgetStatus>("/budgets", data);
    return res.data;
  },
  getById: async (id: string) => {
    const res = await api.get<import("@/types").BudgetStatus>(`/budgets/${id}`);
    return res.data;
  },
  list: async (year?: number, month?: number) => {
    const qs = buildQueryString({ year, month });
    const res = await api.get<{ budgets: import("@/types").BudgetStatus[] }>(
      `/budgets${qs}`
    );
    return res.data.budgets;
  },
  update: async (id: string, data: Partial<import("@/types").CreateBudgetRequest>) => {
    const res = await api.put<import("@/types").BudgetStatus>(
      `/budgets/${id}`,
      data
    );
    return res.data;
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/budgets/${id}`);
  },
};

export const subscriptionService = {
  create: async (data: import("@/types").CreateSubscriptionRequest) => {
    const res = await api.post<import("@/types").Subscription>("/subscriptions", data);
    return res.data;
  },
  list: async (activeOnly = false) => {
    const res = await api.get<{ subscriptions: import("@/types").Subscription[] }>(
      `/subscriptions?active=${activeOnly}`
    );
    return res.data.subscriptions;
  },
  getUpcoming: async (days = 7) => {
    const res = await api.get<{ subscriptions: import("@/types").Subscription[] }>(
      `/subscriptions/upcoming?days=${days}`
    );
    return res.data.subscriptions;
  },
  update: async (id: string, data: Partial<import("@/types").CreateSubscriptionRequest>) => {
    const res = await api.put<import("@/types").Subscription>(
      `/subscriptions/${id}`,
      data
    );
    return res.data;
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/subscriptions/${id}`);
  },
};

export const goalService = {
  create: async (data: import("@/types").CreateGoalRequest) => {
    const res = await api.post<import("@/types").Goal>("/goals", data);
    return res.data;
  },
  list: async (status?: string) => {
    const qs = status ? `?status=${status}` : "";
    const res = await api.get<{ goals: import("@/types").Goal[] }>(`/goals${qs}`);
    return res.data.goals;
  },
  getById: async (id: string) => {
    const res = await api.get<import("@/types").Goal>(`/goals/${id}`);
    return res.data;
  },
  update: async (id: string, data: Partial<import("@/types").CreateGoalRequest>) => {
    const res = await api.put<import("@/types").Goal>(`/goals/${id}`, data);
    return res.data;
  },
  contribute: async (id: string, amount: number) => {
    const res = await api.post<import("@/types").Goal>(`/goals/${id}/contribute`, {
      amount,
    });
    return res.data;
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/goals/${id}`);
  },
};

export const tagService = {
  list: async () => {
    const res = await api.get<{ tags: import("@/types").Tag[] }>("/tags");
    return res.data.tags;
  },
  create: async (name: string, color?: string) => {
    const res = await api.post<import("@/types").Tag>("/tags", { name, color });
    return res.data;
  },
  addToExpense: async (expenseId: string, tagId: string): Promise<void> => {
    await api.post(`/expenses/${expenseId}/tags/${tagId}`);
  },
  removeFromExpense: async (
    expenseId: string,
    tagId: string
  ): Promise<void> => {
    await api.delete(`/expenses/${expenseId}/tags/${tagId}`);
  },
};
