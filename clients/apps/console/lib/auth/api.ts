import { ApiError, api } from "@/lib/api";

export type User = {
  id: string;
  email: string;
  name?: string;
  email_verified: boolean;
};
type Transaction = { transaction_id: string };
type Authentication = { user_id: string; session_expires_at?: string };

export const authApi = {
  start: (email: string) =>
    api<Transaction>("/auth/start", { method: "POST", body: { email } }),
  recover: (email: string) =>
    api<Transaction>("/auth/password/recovery", {
      method: "POST",
      body: { email },
    }),
  sendCode: (transaction_id: string) =>
    api<Transaction>("/auth/otp/send", {
      method: "POST",
      body: { transaction_id },
    }),
  verify: (body: { transaction_id: string; code: string }) =>
    api<Authentication>("/auth/otp/verify", { method: "POST", body }),
  login: (body: { transaction_id: string; password: string }) =>
    api<Authentication>("/auth/password/login", { method: "POST", body }),
  enroll: (password: string) =>
    api<Authentication>("/auth/password/enroll", {
      method: "POST",
      body: { password },
    }),
  reset: (body: { transaction_id: string; code: string; password: string }) =>
    api<void>("/auth/password/reset", { method: "POST", body }),
  logout: () => api<void>("/sessions/current", { method: "DELETE" }),
  async me(signal?: AbortSignal): Promise<User | null> {
    try {
      return await api<User>("/users/me", { signal });
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) return null;
      throw error;
    }
  },
};
