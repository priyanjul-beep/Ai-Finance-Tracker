"use client";

import { useCallback } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { authService, userService } from "@/services/auth.service";
import { useAuthStore } from "@/store/auth.store";
import { queryKeys } from "@/lib/query-client";
import type {
  SignupRequest,
  LoginRequest,
  UpdateUserRequest,
  ChangePasswordRequest,
} from "@/types";

export function useAuth() {
  const router = useRouter();
  const qc = useQueryClient();
  const { user, isAuthenticated, setAuth, clearAuth, updateUser } =
    useAuthStore();

  // ── Profile query (only when authenticated) ────────────────────────────────
  const { data: profile, isLoading: profileLoading } = useQuery({
    queryKey: queryKeys.profile(),
    queryFn: userService.getProfile,
    enabled: isAuthenticated,
    staleTime: 5 * 60 * 1000,
  });

  // ── Sign up ────────────────────────────────────────────────────────────────
  const signupMutation = useMutation({
    mutationFn: (data: SignupRequest) => authService.signup(data),
    onSuccess: (data) => {
      setAuth(data.user, data.access_token, data.refresh_token);
      toast.success("Account created! Please verify your email.");
      router.push("/dashboard");
    },
  });

  // ── Log in ─────────────────────────────────────────────────────────────────
  const loginMutation = useMutation({
    mutationFn: (data: LoginRequest) => authService.login(data),
    onSuccess: (data) => {
      setAuth(data.user, data.access_token, data.refresh_token);
      toast.success(`Welcome back, ${data.user.name}!`);
      router.push("/dashboard");
    },
  });

  // ── Log out ────────────────────────────────────────────────────────────────
  const logoutMutation = useMutation({
    mutationFn: authService.logout,
    onSettled: () => {
      clearAuth();
      qc.clear();
      router.push("/login");
    },
  });

  // ── Update profile ─────────────────────────────────────────────────────────
  const updateProfileMutation = useMutation({
    mutationFn: (data: UpdateUserRequest) => userService.updateProfile(data),
    onSuccess: (updatedUser) => {
      updateUser(updatedUser);
      qc.invalidateQueries({ queryKey: queryKeys.profile() });
      toast.success("Profile updated");
    },
  });

  // ── Change password ────────────────────────────────────────────────────────
  const changePasswordMutation = useMutation({
    mutationFn: (data: ChangePasswordRequest) =>
      authService.changePassword(data),
    onSuccess: () => {
      toast.success("Password changed successfully");
    },
  });

  // ── Forgot password ────────────────────────────────────────────────────────
  const forgotPasswordMutation = useMutation({
    mutationFn: (email: string) => authService.forgotPassword(email),
    onSuccess: () => {
      toast.success(
        "Password reset email sent. Check your inbox."
      );
    },
  });

  const logout = useCallback(() => {
    logoutMutation.mutate();
  }, [logoutMutation]);

  return {
    user: profile ?? user,
    isAuthenticated,
    profileLoading,

    signup: signupMutation.mutate,
    login: loginMutation.mutate,
    logout,
    updateProfile: updateProfileMutation.mutate,
    changePassword: changePasswordMutation.mutate,
    forgotPassword: forgotPasswordMutation.mutate,

    isSigningUp: signupMutation.isPending,
    isLoggingIn: loginMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
    isUpdatingProfile: updateProfileMutation.isPending,
  };
}
