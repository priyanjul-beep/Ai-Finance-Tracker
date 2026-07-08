import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";
import { format, formatDistanceToNow } from "date-fns";

// ─── Tailwind class merging ──────────────────────────────────────────────────
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// ─── Currency formatting ─────────────────────────────────────────────────────
export function formatCurrency(
  amount: number,
  currency = "INR",
  locale = "en-IN"
): string {
  return new Intl.NumberFormat(locale, {
    style: "currency",
    currency,
    maximumFractionDigits: 0,
  }).format(amount);
}

export function formatCompactCurrency(
  amount: number,
  currency = "INR"
): string {
  if (amount >= 100_000) {
    return `${currency === "INR" ? "₹" : "$"}${(amount / 100_000).toFixed(1)}L`;
  }
  if (amount >= 1_000) {
    return `${currency === "INR" ? "₹" : "$"}${(amount / 1_000).toFixed(1)}K`;
  }
  return formatCurrency(amount, currency);
}

// ─── Date formatting ─────────────────────────────────────────────────────────
export function formatDate(date: string | Date): string {
  return format(new Date(date), "dd MMM yyyy");
}

export function formatDateTime(date: string | Date): string {
  return format(new Date(date), "dd MMM yyyy, hh:mm a");
}

export function formatRelativeTime(date: string | Date): string {
  return formatDistanceToNow(new Date(date), { addSuffix: true });
}

export function formatMonth(month: number, year: number): string {
  return format(new Date(year, month - 1, 1), "MMMM yyyy");
}

// ─── Number formatting ───────────────────────────────────────────────────────
export function formatPercent(value: number, decimals = 1): string {
  return `${value.toFixed(decimals)}%`;
}

export function formatNumber(value: number): string {
  return new Intl.NumberFormat("en-IN").format(value);
}

// ─── Category utilities ──────────────────────────────────────────────────────
export const CATEGORY_COLORS: Record<string, string> = {
  food: "#f97316",
  transport: "#3b82f6",
  travel: "#06b6d4",
  shopping: "#a855f7",
  entertainment: "#ec4899",
  health: "#22c55e",
  education: "#eab308",
  utilities: "#64748b",
  subscription: "#8b5cf6",
  investment: "#10b981",
  salary: "#059669",
  freelance: "#0ea5e9",
  business: "#6366f1",
  rental: "#f59e0b",
  others: "#9ca3af",
};

export function getCategoryColor(category: string): string {
  return CATEGORY_COLORS[category.toLowerCase()] ?? "#9ca3af";
}

export const CATEGORY_ICONS: Record<string, string> = {
  food: "🍔",
  transport: "🚗",
  travel: "✈️",
  shopping: "🛍️",
  entertainment: "🎬",
  health: "❤️",
  education: "📚",
  utilities: "💡",
  subscription: "📱",
  investment: "📈",
  salary: "💰",
  freelance: "💻",
  business: "🏢",
  rental: "🏠",
  others: "📦",
};

export function getCategoryIcon(category: string): string {
  return CATEGORY_ICONS[category.toLowerCase()] ?? "📦";
}

// ─── Validation helpers ──────────────────────────────────────────────────────
export function isValidEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function isValidAmount(amount: number): boolean {
  return !isNaN(amount) && amount > 0 && isFinite(amount);
}

// ─── URL helpers ─────────────────────────────────────────────────────────────
export function buildQueryString(params: Record<string, unknown>): string {
  const search = new URLSearchParams();
  Object.entries(params).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== "") {
      search.append(k, String(v));
    }
  });
  return search.toString() ? `?${search.toString()}` : "";
}

// ─── Array / object utilities ────────────────────────────────────────────────
export function groupBy<T>(
  arr: T[],
  key: keyof T
): Record<string, T[]> {
  return arr.reduce<Record<string, T[]>>((acc, item) => {
    const group = String(item[key]);
    if (!acc[group]) acc[group] = [];
    acc[group].push(item);
    return acc;
  }, {});
}

export function sumBy<T>(arr: T[], key: keyof T): number {
  return arr.reduce((acc, item) => acc + Number(item[key]), 0);
}

// ─── Storage helpers ─────────────────────────────────────────────────────────
export const storage = {
  get: (key: string) => {
    if (typeof window === "undefined") return null;
    try {
      return JSON.parse(localStorage.getItem(key) ?? "null");
    } catch {
      return null;
    }
  },
  set: (key: string, value: unknown) => {
    if (typeof window === "undefined") return;
    localStorage.setItem(key, JSON.stringify(value));
  },
  remove: (key: string) => {
    if (typeof window === "undefined") return;
    localStorage.removeItem(key);
  },
};

// ─── Debounce ────────────────────────────────────────────────────────────────
export function debounce<T extends (...args: unknown[]) => unknown>(
  fn: T,
  delay: number
): (...args: Parameters<T>) => void {
  let timer: ReturnType<typeof setTimeout>;
  return (...args) => {
    clearTimeout(timer);
    timer = setTimeout(() => fn(...args), delay);
  };
}
