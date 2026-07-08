import { create } from "zustand";

// ─── UI Theme ─────────────────────────────────────────────────────────────────

type Theme = "light" | "dark" | "system";

// ─── App / UI State ──────────────────────────────────────────────────────────

interface AppState {
  // Sidebar
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  toggleSidebar: () => void;

  // Theme
  theme: Theme;
  setTheme: (theme: Theme) => void;

  // Global loading overlay
  isLoading: boolean;
  setLoading: (loading: boolean) => void;

  // Selected date range (shared across pages)
  dateRange: { from: Date | null; to: Date | null };
  setDateRange: (range: { from: Date | null; to: Date | null }) => void;

  // Selected currency
  currency: string;
  setCurrency: (currency: string) => void;
}

export const useAppStore = create<AppState>((set) => ({
  sidebarOpen: true,
  setSidebarOpen: (open) => set({ sidebarOpen: open }),
  toggleSidebar: () => set((s) => ({ sidebarOpen: !s.sidebarOpen })),

  theme: "system",
  setTheme: (theme) => set({ theme }),

  isLoading: false,
  setLoading: (isLoading) => set({ isLoading }),

  dateRange: { from: null, to: null },
  setDateRange: (dateRange) => set({ dateRange }),

  currency: "INR",
  setCurrency: (currency) => set({ currency }),
}));
