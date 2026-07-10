import api from "./api";
import type {
  Dashboard,
  MonthlyReport,
  PredictionData,
  FinancialHealthScore,
} from "@/types";

export const analyticsService = {
  // Main dashboard data
  getDashboard: async (): Promise<Dashboard> => {
    const res = await api.get<Dashboard>("/analytics/dashboard");
    return res.data;
  },

  // Monthly report
  getMonthlyReport: async (
    year: number,
    month: number
  ): Promise<MonthlyReport> => {
    const res = await api.get<MonthlyReport & { total_expense?: number; total_savings?: number }>(
      `/analytics/monthly/${month}/${year}`
    );
    const d = res.data;
    return {
      ...d,
      // Normalise backend field name variations
      total_expenses: d.total_expenses ?? (d as any).total_expense ?? 0,
      net_savings: d.net_savings ?? (d as any).total_savings ?? 0,
    };
  },

  // Yearly report
  getYearlyReport: async (year: number): Promise<Record<string, unknown>> => {
    const res = await api.get<Record<string, unknown>>(
      `/analytics/yearly?year=${year}`
    );
    return res.data;
  },

  // AI-based expense predictions
  getPredictions: async (): Promise<PredictionData> => {
    const res = await api.get<PredictionData>("/analytics/predictions");
    return res.data;
  },

  // AI-generated insights
  getInsights: async (): Promise<string[]> => {
    const res = await api.get<{ insights: string[] }>("/analytics/insights");
    return res.data.insights;
  },

  // Financial health score
  getHealthScore: async (): Promise<FinancialHealthScore> => {
    const res = await api.get<FinancialHealthScore>(
      "/analytics/health-score"
    );
    return res.data;
  },

  // Export report as PDF or CSV
  exportReport: async (
    type: "pdf" | "csv",
    year: number,
    month?: number
  ): Promise<Blob> => {
    const params = new URLSearchParams({ type, year: String(year) });
    if (month) params.set("month", String(month));
    const res = await api.get(`/analytics/export?${params.toString()}`, {
      responseType: "blob",
    });
    return res.data as Blob;
  },
};
