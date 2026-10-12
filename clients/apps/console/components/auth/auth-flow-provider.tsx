"use client";

import {
  createContext,
  type ReactNode,
  useContext,
  useEffect,
  useState,
} from "react";
import { z } from "zod";

export const flowKey = "leamout.auth-flow";
export const flowSchema = z.object({
  transactionId: z.string().uuid(),
  email: z.string().email(),
  intent: z.enum(["log-in", "create-account", "recovery"]),
  stage: z.enum(["password", "verification", "enroll", "reset"]),
  expiresAt: z.number(),
  resendAt: z.number(),
});
export type AuthFlow = z.infer<typeof flowSchema>;

export function restoreFlow(
  value: string | null,
  now = Date.now(),
): AuthFlow | null {
  try {
    const parsed = flowSchema.safeParse(JSON.parse(value ?? "null"));
    return parsed.success && parsed.data.expiresAt > now ? parsed.data : null;
  } catch {
    return null;
  }
}

const AuthFlowContext = createContext<{
  flow: AuthFlow | null;
  ready: boolean;
  recoveryCode: string;
  setRecoveryCode: (code: string) => void;
  save: (flow: AuthFlow | null) => void;
} | null>(null);

export function AuthFlowProvider({ children }: { children: ReactNode }) {
  const [flow, setFlow] = useState<AuthFlow | null>(null);
  const [ready, setReady] = useState(false);
  const [recoveryCode, setRecoveryCode] = useState("");

  useEffect(() => {
    try {
      setFlow(restoreFlow(sessionStorage.getItem(flowKey)));
    } catch {
      // The flow still works in memory when browser storage is unavailable.
    }
    setReady(true);
  }, []);

  function save(next: AuthFlow | null) {
    setFlow(next);
    if (!next || next.transactionId !== flow?.transactionId)
      setRecoveryCode("");
    try {
      if (next) sessionStorage.setItem(flowKey, JSON.stringify(next));
      else sessionStorage.removeItem(flowKey);
    } catch {
      // Storage is optional; never persist passwords or verification codes.
    }
  }

  return (
    <AuthFlowContext.Provider
      value={{ flow, ready, recoveryCode, setRecoveryCode, save }}
    >
      {children}
    </AuthFlowContext.Provider>
  );
}

export function useAuthFlow() {
  const value = useContext(AuthFlowContext);
  if (!value) throw new Error("AuthFlowProvider is required");
  return value;
}
