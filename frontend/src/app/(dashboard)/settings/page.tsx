"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import {
  User, Lock, Bell, Trash2, Loader2, Eye, EyeOff, Camera,
} from "lucide-react";
import { useAuth } from "@/hooks/useAuth";
import type { UpdateUserRequest, ChangePasswordRequest } from "@/types";

const TIMEZONES = [
  "Asia/Kolkata", "Asia/Dubai", "Asia/Singapore", "Asia/Tokyo",
  "Europe/London", "Europe/Paris", "America/New_York", "America/Los_Angeles",
  "Australia/Sydney", "UTC",
];

const CURRENCIES = ["INR", "USD", "EUR", "GBP", "AED", "SGD", "AUD", "JPY"];
const LANGUAGES  = [
  { label: "English", value: "en" },
  { label: "Hindi",   value: "hi" },
  { label: "Tamil",   value: "ta" },
  { label: "Telugu",  value: "te" },
];

type Tab = "profile" | "security" | "preferences";

export default function SettingsPage() {
  const [activeTab, setActiveTab] = useState<Tab>("profile");
  const [showCurrentPw, setShowCurrentPw] = useState(false);
  const [showNewPw,     setShowNewPw]     = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const { user, updateProfile, changePassword, isUpdatingProfile } = useAuth();

  // ── Profile form ──────────────────────────────────────────────────────────
  const profileForm = useForm<UpdateUserRequest>({
    defaultValues: {
      name:               user?.name               ?? "",
      timezone:           user?.timezone            ?? "Asia/Kolkata",
      currency:           user?.currency            ?? "INR",
      preferred_language: user?.preferred_language  ?? "en",
    },
  });

  // ── Password form ─────────────────────────────────────────────────────────
  const pwForm = useForm<ChangePasswordRequest & { confirm_password: string }>({});
  const [pwError, setPwError] = useState("");
  const [pwLoading, setPwLoading] = useState(false);

  const onProfileSubmit = (data: UpdateUserRequest) => {
    updateProfile(data);
  };

  const onPasswordSubmit = (data: ChangePasswordRequest & { confirm_password: string }) => {
    setPwError("");
    if (data.new_password !== data.confirm_password) {
      setPwError("Passwords do not match");
      return;
    }
    setPwLoading(true);
    changePassword(
      { current_password: data.current_password, new_password: data.new_password },
      { onSettled: () => setPwLoading(false) }
    );
  };

  const tabs: { id: Tab; label: string; icon: React.ReactNode }[] = [
    { id: "profile",     label: "Profile",      icon: <User  className="h-4 w-4" /> },
    { id: "security",    label: "Security",      icon: <Lock  className="h-4 w-4" /> },
    { id: "preferences", label: "Preferences",  icon: <Bell  className="h-4 w-4" /> },
  ];

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-sm text-muted-foreground">Manage your account and preferences</p>
      </div>

      {/* Tab bar */}
      <div className="flex gap-1 rounded-xl border border-border bg-muted/40 p-1">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setActiveTab(t.id)}
            className={`flex flex-1 items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
              activeTab === t.id
                ? "bg-background shadow-sm text-foreground"
                : "text-muted-foreground hover:text-foreground"
            }`}
          >
            {t.icon}
            {t.label}
          </button>
        ))}
      </div>

      {/* ── Profile tab ──────────────────────────────────────────────────── */}
      {activeTab === "profile" && (
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-6">
          {/* Avatar */}
          <div className="flex items-center gap-4">
            <div className="relative">
              <div className="h-16 w-16 rounded-full bg-primary/20 flex items-center justify-center text-2xl font-bold text-primary">
                {user?.name?.charAt(0).toUpperCase() ?? "U"}
              </div>
              <button className="absolute -bottom-1 -right-1 flex h-6 w-6 items-center justify-center rounded-full border border-border bg-background shadow-sm hover:bg-muted transition-colors">
                <Camera className="h-3 w-3" />
              </button>
            </div>
            <div>
              <p className="font-semibold">{user?.name}</p>
              <p className="text-sm text-muted-foreground">{user?.email}</p>
              {user?.is_email_verified ? (
                <span className="mt-1 inline-flex items-center rounded-full bg-green-100 px-2 py-0.5 text-xs font-medium text-green-700">
                  Email verified
                </span>
              ) : (
                <span className="mt-1 inline-flex items-center rounded-full bg-yellow-100 px-2 py-0.5 text-xs font-medium text-yellow-700">
                  Email not verified
                </span>
              )}
            </div>
          </div>

          {/* Profile form */}
          <form onSubmit={profileForm.handleSubmit(onProfileSubmit)} className="space-y-4">
            <div className="space-y-1.5">
              <label className="text-sm font-medium">Full Name</label>
              <input
                type="text"
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                {...profileForm.register("name", { required: true })}
              />
            </div>

            <div className="space-y-1.5">
              <label className="text-sm font-medium">Email</label>
              <input
                type="email"
                value={user?.email ?? ""}
                disabled
                className="w-full rounded-lg border border-input bg-muted/50 px-3 py-2 text-sm text-muted-foreground cursor-not-allowed"
              />
            </div>

            <button
              type="submit"
              disabled={isUpdatingProfile}
              className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60 transition-colors"
            >
              {isUpdatingProfile && <Loader2 className="h-4 w-4 animate-spin" />}
              Save Changes
            </button>
          </form>
        </div>
      )}

      {/* ── Security tab ─────────────────────────────────────────────────── */}
      {activeTab === "security" && (
        <div className="space-y-4">
          <div className="rounded-xl border border-border bg-card p-6 shadow-sm space-y-4">
            <h2 className="text-sm font-semibold">Change Password</h2>

            <form onSubmit={pwForm.handleSubmit(onPasswordSubmit)} className="space-y-4">
              {/* Current password */}
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Current Password</label>
                <div className="relative">
                  <input
                    type={showCurrentPw ? "text" : "password"}
                    placeholder="••••••••"
                    className="w-full rounded-lg border border-input bg-background px-3 py-2 pr-10 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                    {...pwForm.register("current_password", { required: true })}
                  />
                  <button
                    type="button"
                    onClick={() => setShowCurrentPw((v) => !v)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showCurrentPw ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>

              {/* New password */}
              <div className="space-y-1.5">
                <label className="text-sm font-medium">New Password</label>
                <div className="relative">
                  <input
                    type={showNewPw ? "text" : "password"}
                    placeholder="Min 8 characters"
                    className="w-full rounded-lg border border-input bg-background px-3 py-2 pr-10 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                    {...pwForm.register("new_password", { required: true, minLength: 8 })}
                  />
                  <button
                    type="button"
                    onClick={() => setShowNewPw((v) => !v)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                  >
                    {showNewPw ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </div>

              {/* Confirm password */}
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Confirm New Password</label>
                <input
                  type="password"
                  placeholder="Repeat new password"
                  className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                  {...pwForm.register("confirm_password", { required: true })}
                />
              </div>

              {pwError && <p className="text-xs text-destructive">{pwError}</p>}

              <button
                type="submit"
                disabled={pwLoading}
                className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60 transition-colors"
              >
                {pwLoading && <Loader2 className="h-4 w-4 animate-spin" />}
                Update Password
              </button>
            </form>
          </div>

          {/* Danger zone */}
          <div className="rounded-xl border border-destructive/40 bg-card p-6 shadow-sm space-y-3">
            <h2 className="text-sm font-semibold text-destructive">Danger Zone</h2>
            <p className="text-sm text-muted-foreground">
              Permanently delete your account and all associated data. This action cannot be undone.
            </p>
            {!confirmDelete ? (
              <button
                onClick={() => setConfirmDelete(true)}
                className="flex items-center gap-2 rounded-lg border border-destructive px-4 py-2 text-sm font-medium text-destructive hover:bg-destructive/10 transition-colors"
              >
                <Trash2 className="h-4 w-4" />
                Delete Account
              </button>
            ) : (
              <div className="space-y-3">
                <p className="text-sm font-medium text-destructive">
                  Are you absolutely sure? This cannot be reversed.
                </p>
                <div className="flex gap-3">
                  <button
                    onClick={() => setConfirmDelete(false)}
                    className="rounded-lg border border-border px-4 py-2 text-sm font-medium hover:bg-muted transition-colors"
                  >
                    Cancel
                  </button>
                  <button className="rounded-lg bg-destructive px-4 py-2 text-sm font-medium text-destructive-foreground hover:bg-destructive/90 transition-colors">
                    Yes, delete my account
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {/* ── Preferences tab ───────────────────────────────────────────────── */}
      {activeTab === "preferences" && (
        <div className="rounded-xl border border-border bg-card p-6 shadow-sm">
          <form onSubmit={profileForm.handleSubmit(onProfileSubmit)} className="space-y-5">
            <h2 className="text-sm font-semibold">Regional & Display</h2>

            <div className="space-y-1.5">
              <label className="text-sm font-medium">Timezone</label>
              <select
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                {...profileForm.register("timezone")}
              >
                {TIMEZONES.map((tz) => (
                  <option key={tz} value={tz}>{tz}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="text-sm font-medium">Default Currency</label>
              <select
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                {...profileForm.register("currency")}
              >
                {CURRENCIES.map((c) => (
                  <option key={c} value={c}>{c}</option>
                ))}
              </select>
            </div>

            <div className="space-y-1.5">
              <label className="text-sm font-medium">Language</label>
              <select
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-1 focus:ring-primary"
                {...profileForm.register("preferred_language")}
              >
                {LANGUAGES.map((l) => (
                  <option key={l.value} value={l.value}>{l.label}</option>
                ))}
              </select>
            </div>

            <button
              type="submit"
              disabled={isUpdatingProfile}
              className="flex items-center gap-2 rounded-lg bg-primary px-4 py-2 text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60 transition-colors"
            >
              {isUpdatingProfile && <Loader2 className="h-4 w-4 animate-spin" />}
              Save Preferences
            </button>
          </form>
        </div>
      )}
    </div>
  );
}
