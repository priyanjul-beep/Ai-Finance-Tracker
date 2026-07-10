"use client";

import { useState, useEffect, useRef, useCallback } from "react";
import {
  Bell,
  Check,
  CheckCheck,
  Trash2,
  Loader2,
  Info,
  AlertTriangle,
  AlertOctagon,
  DollarSign,
  Target,
  Gift,
  Calendar,
  Filter,
} from "lucide-react";
import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from "@tanstack/react-query";
import { api } from "@/services/api";
import { toast } from "sonner";
import type { Notification, NotificationListResponse, NotificationType } from "@/types";

// ─── Service ──────────────────────────────────────────────────────────────────

const notificationService = {
  list: async (page = 1, limit = 20, type = ""): Promise<NotificationListResponse> => {
    const params = new URLSearchParams({ page: String(page), limit: String(limit) });
    if (type) params.set("type", type);
    const res = await api.get(`/notifications?${params}`);
    return res.data;
  },
  markRead: async (id: string): Promise<void> => {
    await api.patch(`/notifications/${id}/read`);
  },
  markAllRead: async (): Promise<void> => {
    await api.patch("/notifications/read-all");
  },
  delete: async (id: string): Promise<void> => {
    await api.delete(`/notifications/${id}`);
  },
};

// ─── Helpers ─────────────────────────────────────────────────────────────────

function relativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const s = Math.floor(diff / 1000);
  if (s < 60) return "just now";
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.floor(h / 24);
  if (d < 7) return `${d}d ago`;
  return new Date(dateStr).toLocaleDateString();
}

function PriorityBadge({ priority }: { priority: string }) {
  const map: Record<string, { label: string; cls: string }> = {
    critical: { label: "Critical", cls: "bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400" },
    high:     { label: "High",     cls: "bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400" },
    medium:   { label: "Medium",   cls: "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400" },
    low:      { label: "Low",      cls: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400" },
  };
  const p = map[priority] ?? map.low;
  return (
    <span className={`inline-flex items-center rounded-full px-1.5 py-0.5 text-[10px] font-semibold ${p.cls}`}>
      {p.label}
    </span>
  );
}

function NotifIcon({ type }: { type: string }) {
  const cls = "h-4 w-4 flex-shrink-0";
  if (type === "budget_exceeded") return <AlertOctagon className={`${cls} text-red-500`} />;
  if (type === "budget_warning" || type === "budget_alert") return <AlertTriangle className={`${cls} text-yellow-500`} />;
  if (type === "goal_achievement") return <Target className={`${cls} text-blue-500`} />;
  if (type?.includes("subscription")) return <DollarSign className={`${cls} text-purple-500`} />;
  if (type === "welcome") return <Gift className={`${cls} text-green-500`} />;
  if (type?.includes("summary")) return <Calendar className={`${cls} text-indigo-500`} />;
  return <Info className={`${cls} text-primary`} />;
}

function TypeDot({ type }: { type: string }) {
  const map: Record<string, string> = {
    budget_exceeded: "bg-red-500",
    budget_warning:  "bg-yellow-500",
    budget_alert:    "bg-yellow-500",
    goal_achievement:"bg-blue-500",
    welcome:         "bg-green-500",
    monthly_summary: "bg-indigo-500",
    weekly_summary:  "bg-indigo-500",
  };
  return <span className={`h-2 w-2 rounded-full flex-shrink-0 ${map[type] ?? "bg-gray-400"}`} />;
}

// ─── Filter tabs ─────────────────────────────────────────────────────────────

const TYPE_FILTERS: { label: string; value: string }[] = [
  { label: "All", value: "" },
  { label: "Budgets", value: "budget_warning" },
  { label: "Exceeded", value: "budget_exceeded" },
  { label: "Goals", value: "goal_achievement" },
  { label: "Welcome", value: "welcome" },
  { label: "Summaries", value: "monthly_summary" },
];

// ─── Page ────────────────────────────────────────────────────────────────────

const PAGE_LIMIT = 20;

export default function NotificationsPage() {
  const qc = useQueryClient();
  const [activeType, setActiveType] = useState("");
  const loadMoreRef = useRef<HTMLDivElement>(null);

  // Infinite scroll query
  const {
    data,
    isLoading,
    isFetchingNextPage,
    fetchNextPage,
    hasNextPage,
  } = useInfiniteQuery({
    queryKey: ["notifications", "infinite", activeType],
    queryFn: ({ pageParam = 1 }) =>
      notificationService.list(pageParam as number, PAGE_LIMIT, activeType),
    getNextPageParam: (last, pages) => {
      const loaded = pages.length * PAGE_LIMIT;
      return loaded < last.total ? pages.length + 1 : undefined;
    },
    initialPageParam: 1,
  });

  // Unread count from first page
  const unreadCount = data?.pages[0]?.unread_count ?? 0;
  const allNotifications: Notification[] = data?.pages.flatMap((p) => p.notifications) ?? [];
  const total = data?.pages[0]?.total ?? 0;

  // Intersection observer for infinite scroll
  const onIntersect = useCallback(
    (entries: IntersectionObserverEntry[]) => {
      if (entries[0].isIntersecting && hasNextPage && !isFetchingNextPage) {
        fetchNextPage();
      }
    },
    [hasNextPage, isFetchingNextPage, fetchNextPage]
  );

  useEffect(() => {
    const el = loadMoreRef.current;
    if (!el) return;
    const obs = new IntersectionObserver(onIntersect, { threshold: 0.5 });
    obs.observe(el);
    return () => obs.disconnect();
  }, [onIntersect]);

  const { mutate: markRead } = useMutation({
    mutationFn: notificationService.markRead,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
    },
  });

  const { mutate: markAllRead, isPending: isMarkingAll } = useMutation({
    mutationFn: notificationService.markAllRead,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      toast.success("All notifications marked as read");
    },
  });

  const { mutate: deleteNotif } = useMutation({
    mutationFn: notificationService.delete,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      toast.success("Notification deleted");
    },
  });

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Notifications</h1>
          <p className="text-sm text-muted-foreground">
            {total} notification{total !== 1 ? "s" : ""}
            {unreadCount > 0 && (
              <span className="ml-2 inline-flex items-center rounded-full bg-destructive/10 px-2 py-0.5 text-xs font-semibold text-destructive">
                {unreadCount} unread
              </span>
            )}
          </p>
        </div>
        {unreadCount > 0 && (
          <button
            onClick={() => markAllRead()}
            disabled={isMarkingAll}
            className="flex items-center gap-2 rounded-lg border border-border px-3 py-1.5 text-sm font-medium hover:bg-muted disabled:opacity-60 transition-colors"
          >
            {isMarkingAll ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <CheckCheck className="h-4 w-4" />
            )}
            Mark all read
          </button>
        )}
      </div>

      {/* Type filter tabs */}
      <div className="flex items-center gap-1 overflow-x-auto pb-1">
        <Filter className="h-3.5 w-3.5 flex-shrink-0 text-muted-foreground mr-1" />
        {TYPE_FILTERS.map((f) => (
          <button
            key={f.value}
            onClick={() => setActiveType(f.value)}
            className={`flex-shrink-0 rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              activeType === f.value
                ? "bg-primary text-primary-foreground"
                : "border border-border hover:bg-muted text-muted-foreground"
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      {/* Notification list */}
      <div className="rounded-xl border border-border bg-card shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : allNotifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Bell className="h-10 w-10 text-muted-foreground mb-3" />
            <p className="text-sm font-medium text-muted-foreground">You're all caught up!</p>
            <p className="text-xs text-muted-foreground mt-1">
              {activeType ? "No notifications of this type" : "No notifications yet"}
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-border">
            {allNotifications.map((notif) => (
              <li
                key={notif.id}
                className={`group flex items-start gap-3 px-4 py-3.5 transition-colors hover:bg-muted/30 ${
                  notif.is_read ? "bg-background" : "bg-primary/5"
                }`}
              >
                {/* Unread dot */}
                <div className="mt-1.5 flex-shrink-0">
                  {notif.is_read ? (
                    <NotifIcon type={notif.type} />
                  ) : (
                    <div className="relative">
                      <NotifIcon type={notif.type} />
                      <span className="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-primary ring-2 ring-background" />
                    </div>
                  )}
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 flex-wrap">
                    <p className={`text-sm ${notif.is_read ? "font-normal" : "font-semibold"}`}>
                      {notif.title}
                    </p>
                    <PriorityBadge priority={notif.priority} />
                    <TypeDot type={notif.type} />
                  </div>
                  <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                    {notif.message}
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-1">
                    {relativeTime(notif.created_at)}
                  </p>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-1 flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
                  {!notif.is_read && (
                    <button
                      onClick={() => markRead(notif.id)}
                      className="rounded p-1.5 hover:bg-muted transition-colors text-muted-foreground hover:text-foreground"
                      title="Mark as read"
                    >
                      <Check className="h-3.5 w-3.5" />
                    </button>
                  )}
                  <button
                    onClick={() => deleteNotif(notif.id)}
                    className="rounded p-1.5 hover:bg-destructive/10 transition-colors text-muted-foreground hover:text-destructive"
                    title="Delete"
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}

        {/* Infinite scroll sentinel */}
        <div ref={loadMoreRef} className="py-2 flex justify-center">
          {isFetchingNextPage && (
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>
      </div>
    </div>
  );
}
