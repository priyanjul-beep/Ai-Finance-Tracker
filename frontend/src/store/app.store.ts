import { create } from "zustand";

// ─── UI Theme ─────────────────────────────────────────────────────────────────

type Theme = "light" | "dark" | "system";

// ─── App / UI State ──────────────────────────────────────────────────────────

interface AppState {
  // Sidebar (desktop collapse)
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  toggleSidebar: () => void;

  // Mobile sidebar drawer
  mobileSidebarOpen: boolean;
  setMobileSidebarOpen: (open: boolean) => void;
  toggleMobileSidebar: () => void;

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

  mobileSidebarOpen: false,
  setMobileSidebarOpen: (open) => set({ mobileSidebarOpen: open }),
  toggleMobileSidebar: () => set((s) => ({ mobileSidebarOpen: !s.mobileSidebarOpen })),

  theme: "system",
  setTheme: (theme) => set({ theme }),

  isLoading: false,
  setLoading: (isLoading) => set({ isLoading }),

  dateRange: { from: null, to: null },
  setDateRange: (dateRange) => set({ dateRange }),

  currency: "INR",
  setCurrency: (currency) => set({ currency }),
}));
