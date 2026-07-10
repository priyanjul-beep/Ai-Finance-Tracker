"use client";

import { Bell, Plus, Menu } from "lucide-react";
import Link from "next/link";
import { useAuth } from "@/hooks/useAuth";
import { useAppStore } from "@/store/app.store";
import { formatDate } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { api } from "@/services/api";

interface HeaderProps {
  title?: string;
}

async function fetchUnreadCount(): Promise<number> {
  try {
    const res = await api.get("/notifications/unread-count");
    return res.data?.count ?? 0;
  } catch {
    return 0;
  }
}

export function Header({ title }: HeaderProps) {
  const { user } = useAuth();
  const { toggleMobileSidebar } = useAppStore();
  const today = formatDate(new Date());

  const { data: unreadCount = 0 } = useQuery({
    queryKey: ["notifications", "unread-count"],
    queryFn: fetchUnreadCount,
    refetchInterval: 30_000, // poll every 30 s
    staleTime: 20_000,
  });

  return (
    <header className="flex h-16 items-center justify-between border-b border-border bg-background px-4 md:px-6">
      {/* Left: hamburger (mobile) + date */}
      <div className="flex items-center gap-3">
        <button
          onClick={toggleMobileSidebar}
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors md:hidden"
          aria-label="Open menu"
        >
          <Menu className="h-4 w-4" />
        </button>
        <div>
          {title && (
            <h1 className="text-lg font-semibold text-foreground">{title}</h1>
          )}
          <p className="text-xs text-muted-foreground">{today}</p>
        </div>
      </div>

      {/* Right actions */}
      <div className="flex items-center gap-2 md:gap-3">
        {/* Quick add button */}
        <Link
          href="/expenses/new"
          className="flex items-center gap-1.5 rounded-lg bg-primary px-3 py-2 text-xs font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="h-3.5 w-3.5" />
          <span className="hidden sm:inline">Add Expense</span>
        </Link>

        {/* Notifications bell */}
        <Link
          href="/notifications"
          className="relative flex h-9 w-9 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors"
          aria-label={`Notifications${unreadCount > 0 ? ` (${unreadCount} unread)` : ""}`}
        >
          <Bell className="h-4 w-4" />
          {unreadCount > 0 && (
            <span className="absolute -right-0.5 -top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-[10px] font-bold text-white">
              {unreadCount > 99 ? "99+" : unreadCount}
            </span>
          )}
        </Link>

        {/* Avatar */}
        <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center text-xs font-bold text-primary cursor-pointer">
          {user?.name?.charAt(0).toUpperCase() ?? "U"}
        </div>
      </div>
    </header>
  );
}

