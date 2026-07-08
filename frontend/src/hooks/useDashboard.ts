"use client";

import { useQuery } from "@tanstack/react-query";
import { analyticsService } from "@/services/analytics.service";
import { queryKeys } from "@/lib/query-client";

export function useDashboard() {
  return useQuery({
    queryKey: queryKeys.analytics.dashboard(),
    queryFn: analyticsService.getDashboard,
    staleTime: 2 * 60 * 1000, // 2 minutes
    refetchInterval: 5 * 60 * 1000, // refresh every 5 minutes
  });
}

export function useMonthlyReport(year: number, month: number) {
  return useQuery({
    queryKey: queryKeys.analytics.monthly(year, month),
    queryFn: () => analyticsService.getMonthlyReport(year, month),
    staleTime: 5 * 60 * 1000,
    enabled: !!year && !!month,
  });
}

export function useYearlyReport(year: number) {
  return useQuery({
    queryKey: queryKeys.analytics.yearly(year),
    queryFn: () => analyticsService.getYearlyReport(year),
    staleTime: 10 * 60 * 1000,
    enabled: !!year,
  });
}

export function usePredictions() {
  return useQuery({
    queryKey: queryKeys.analytics.predictions(),
    queryFn: analyticsService.getPredictions,
    staleTime: 10 * 60 * 1000,
  });
}

export function useInsights() {
  return useQuery({
    queryKey: queryKeys.analytics.insights(),
    queryFn: analyticsService.getInsights,
    staleTime: 15 * 60 * 1000,
  });
}

export function useHealthScore() {
  return useQuery({
    queryKey: queryKeys.analytics.healthScore(),
    queryFn: analyticsService.getHealthScore,
    staleTime: 10 * 60 * 1000,
  });
}
