"use client";

import { Bell, Search, Plus } from "lucide-react";
import Link from "next/link";
import { useAuth } from "@/hooks/useAuth";
import { formatDate } from "@/lib/utils";

interface HeaderProps {
  title?: string;
}

export function Header({ title }: HeaderProps) {
  const { user } = useAuth();
  const today = formatDate(new Date());

  return (
    <header className="flex h-16 items-center justify-between border-b border-border bg-background px-6">
      {/* Title */}
      <div>
        {title && (
          <h1 className="text-lg font-semibold text-foreground">{title}</h1>
        )}
        <p className="text-xs text-muted-foreground">{today}</p>
      </div>

      {/* Right actions */}
      <div className="flex items-center gap-3">
        {/* Global search */}
        <div className="relative hidden md:flex items-center">
          <Search className="absolute left-3 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search expenses…"
            className="h-9 w-64 rounded-lg border border-input bg-muted/50 pl-9 pr-4 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
          />
        </div>

        {/* Quick add button */}
        <Link
          href="/expenses/new"
          className="flex items-center gap-2 rounded-lg bg-primary px-3 py-2 text-xs font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
        >
          <Plus className="h-3.5 w-3.5" />
          Add Expense
        </Link>

        {/* Notifications */}
        <Link
          href="/notifications"
          className="relative flex h-9 w-9 items-center justify-center rounded-lg border border-border hover:bg-muted transition-colors"
        >
          <Bell className="h-4 w-4" />
          {/* Unread badge */}
          <span className="absolute -right-0.5 -top-0.5 flex h-4 w-4 items-center justify-center rounded-full bg-destructive text-[10px] font-bold text-white">
            3
          </span>
        </Link>

        {/* Avatar */}
        <div className="h-8 w-8 rounded-full bg-primary/20 flex items-center justify-center text-xs font-bold text-primary cursor-pointer">
          {user?.name?.charAt(0).toUpperCase() ?? "U"}
        </div>
      </div>
    </header>
  );
}
