import api from "./api";
import type {
  AuthResponse,
  SignupRequest,
  LoginRequest,
  User,
  UpdateUserRequest,
  ChangePasswordRequest,
} from "@/types";

export const authService = {
  // Sign up a new user
  signup: async (data: SignupRequest): Promise<AuthResponse> => {
    const res = await api.post<AuthResponse>("/auth/signup", data);
    return res.data;
  },

  // Log in an existing user
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const res = await api.post<AuthResponse>("/auth/login", data);
    return res.data;
  },

  // Refresh access token
  refresh: async (refreshToken: string): Promise<AuthResponse> => {
    const res = await api.post<AuthResponse>("/auth/refresh", {
      refresh_token: refreshToken,
    });
    return res.data;
  },

  // Verify email address
  verifyEmail: async (token: string): Promise<void> => {
    await api.get(`/auth/verify-email?token=${token}`);
  },

  // Initiate forgot-password flow
  forgotPassword: async (email: string): Promise<void> => {
    await api.post("/auth/forgot-password", { email });
  },

  // Complete password reset
  resetPassword: async (
    token: string,
    newPassword: string
  ): Promise<void> => {
    await api.post("/auth/reset-password", { token, new_password: newPassword });
  },

  // Change password (authenticated)
  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    await api.post("/auth/change-password", data);
  },

  // Log out (revoke all sessions)
  logout: async (): Promise<void> => {
    await api.post("/auth/logout");
  },

  // Get Google OAuth URL
  getGoogleOAuthURL: (): string => {
    const clientId = process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID;
    const redirectUri = `${process.env.NEXT_PUBLIC_API_URL}/api/v1/auth/google/callback`;
    const scope = encodeURIComponent(
      "openid email profile"
    );
    return `https://accounts.google.com/o/oauth2/v2/auth?client_id=${clientId}&redirect_uri=${redirectUri}&response_type=code&scope=${scope}`;
  },
};

export const userService = {
  getProfile: async (): Promise<User> => {
    const res = await api.get<User>("/user/profile");
    return res.data;
  },

  updateProfile: async (data: UpdateUserRequest): Promise<User> => {
    const res = await api.put<User>("/user/profile", data);
    return res.data;
  },

  uploadAvatar: async (file: File): Promise<{ url: string }> => {
    const form = new FormData();
    form.append("avatar", file);
    const res = await api.post<{ url: string }>("/user/avatar", form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return res.data;
  },

  deleteAccount: async (): Promise<void> => {
    await api.delete("/user/account");
  },
};
