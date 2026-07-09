"use client";

import { useState } from "react";
import { Bell, Check, CheckCheck, Trash2, Loader2, Info, AlertTriangle, DollarSign, Target } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api } from "@/services/api";
import { toast } from "sonner";
import type { Notification } from "@/types";

// ─── Service ──────────────────────────────────────────────────────────────────

const notificationService = {
  list: async (page = 1, limit = 20): Promise<{ data: Notification[]; total: number }> => {
    const res = await api.get(`/notifications?page=${page}&limit=${limit}`);
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

// ─── Icon by type ─────────────────────────────────────────────────────────────

function NotifIcon({ type }: { type: string }) {
  const cls = "h-4 w-4 flex-shrink-0";
  if (type?.includes("budget"))       return <AlertTriangle className={`${cls} text-yellow-500`} />;
  if (type?.includes("goal"))         return <Target        className={`${cls} text-blue-500`}   />;
  if (type?.includes("subscription")) return <DollarSign    className={`${cls} text-purple-500`} />;
  return                                     <Info          className={`${cls} text-primary`}    />;
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export default function NotificationsPage() {
  const [page, setPage] = useState(1);
  const qc = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ["notifications", page],
    queryFn: () => notificationService.list(page, 20),
  });

  const notifications: Notification[] = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / 20));
  const unreadCount = notifications.filter((n) => !n.is_read).length;

  const { mutate: markRead } = useMutation({
    mutationFn: notificationService.markRead,
    onSuccess: () => qc.invalidateQueries({ queryKey: ["notifications"] }),
  });

  const { mutate: markAllRead, isPending: isMarkingAll } = useMutation({
    mutationFn: notificationService.markAllRead,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["notifications"] });
      toast.success("All notifications marked as read");
    },
  });

  const { mutate: deleteNotif, isPending: isDeleting } = useMutation({
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
            {unreadCount > 0 && ` · ${unreadCount} unread`}
          </p>
        </div>
        {unreadCount > 0 && (
          <button
            onClick={() => markAllRead()}
            disabled={isMarkingAll}
            className="flex items-center gap-2 rounded-lg border border-border px-3 py-1.5 text-sm font-medium hover:bg-muted disabled:opacity-60 transition-colors"
          >
            {isMarkingAll
              ? <Loader2 className="h-4 w-4 animate-spin" />
              : <CheckCheck className="h-4 w-4" />
            }
            Mark all read
          </button>
        )}
      </div>

      {/* List */}
      <div className="rounded-xl border border-border bg-card shadow-sm overflow-hidden">
        {isLoading ? (
          <div className="flex items-center justify-center py-20">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : notifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Bell className="h-10 w-10 text-muted-foreground mb-3" />
            <p className="text-sm font-medium text-muted-foreground">
              You're all caught up!
            </p>
            <p className="text-xs text-muted-foreground mt-1">
              No notifications yet
            </p>
          </div>
        ) : (
          <ul className="divide-y divide-border">
            {notifications.map((notif) => (
              <li
                key={notif.id}
                className={`flex items-start gap-3 px-4 py-3.5 transition-colors ${
                  notif.is_read ? "bg-background" : "bg-primary/5"
                }`}
              >
                {/* Type icon */}
                <div className="mt-0.5">
                  <NotifIcon type={notif.type} />
                </div>

                {/* Content */}
                <div className="flex-1 min-w-0">
                  <p className={`text-sm ${notif.is_read ? "font-normal" : "font-semibold"}`}>
                    {notif.title}
                  </p>
                  <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                    {notif.message}
                  </p>
                  <p className="text-xs text-muted-foreground/60 mt-1">
                    {new Date(notif.created_at).toLocaleString()}
                  </p>
                </div>

                {/* Actions */}
                <div className="flex items-center gap-1 flex-shrink-0">
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
                    disabled={isDeleting}
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
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between text-sm">
          <p className="text-muted-foreground">Page {page} of {totalPages}</p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted disabled:opacity-40 transition-colors text-xs"
            >
              ‹
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="flex h-8 w-8 items-center justify-center rounded-lg border border-border hover:bg-muted disabled:opacity-40 transition-colors text-xs"
            >
              ›
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
