"use client";

import { useEffect } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";
import { useAuthStore } from "@/store/auth.store";
import { userService } from "@/services/auth.service";

export default function OAuthCallbackPage() {
  const router = useRouter();
  const params = useSearchParams();
  const { setAuth, clearAuth } = useAuthStore();

  useEffect(() => {
    const accessToken  = params.get("access_token");
    const refreshToken = params.get("refresh_token");
    const error        = params.get("error");

    if (error || !accessToken || !refreshToken) {
      toast.error("Google sign-in failed. Please try again.");
      router.replace("/login");
      return;
    }

    // Temporarily store tokens so the profile API call is authenticated
    // We need to fetch the user profile to populate the auth store
    async function finish() {
      try {
        // Put a temp entry so axios interceptor sends the Authorization header
        useAuthStore.getState().setAuth(
          { id: "", name: "", email: "", is_email_verified: false, timezone: "", currency: "INR", preferred_language: "en", created_at: "", updated_at: "" },
          accessToken!,
          refreshToken!,
        );

        // Fetch real profile now that the token is in the store
        const profile = await userService.getProfile();
        setAuth(profile, accessToken!, refreshToken!);
        toast.success(`Welcome, ${profile.name}!`);
        router.replace("/dashboard");
      } catch {
        clearAuth();
        toast.error("Could not load profile. Please try again.");
        router.replace("/login");
      }
    }

    finish();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-primary/5 via-background to-primary/10">
      <div className="flex flex-col items-center gap-3 text-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
        <p className="text-sm font-medium text-muted-foreground">
          Signing you in with Google…
        </p>
      </div>
    </div>
  );
}
