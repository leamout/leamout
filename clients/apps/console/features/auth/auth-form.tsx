"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { authRequest } from "./api";

type Mode =
  | "login"
  | "login-password"
  | "signup"
  | "verify"
  | "enroll"
  | "recovery"
  | "reset";
type Transaction = {
  id: string;
  email: string;
  purpose: "login" | "signup" | "recovery";
  step: string;
  resendAt: number;
};
const key = "leamout.auth.transaction";
const titles: Record<Mode, string> = {
  login: "Log in to Leamout",
  "login-password": "Enter your password",
  signup: "Create your account",
  verify: "Check your email",
  enroll: "Create a password",
  recovery: "Forgot your password?",
  reset: "Set a new password",
};

export function AuthForm({ mode }: { mode: Mode }) {
  const router = useRouter();
  const [transaction, setTransaction] = useState<Transaction | null>(null);
  const [ready, setReady] = useState(false);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const [code, setCode] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [notice, setNotice] = useState("");
  const [now, setNow] = useState(Date.now());
  const entry = ["login", "signup", "recovery"].includes(mode);
  const settingPassword = mode === "enroll" || mode === "reset";
  const restart =
    mode === "enroll" || transaction?.purpose === "signup"
      ? "/create-account"
      : mode === "reset" || transaction?.purpose === "recovery"
        ? "/forgot-password"
        : "/log-in";

  useEffect(() => {
    try {
      const saved = sessionStorage.getItem(key);
      if (saved) {
        const value = JSON.parse(saved) as Transaction;
        if (typeof value.id === "string" && typeof value.email === "string")
          setTransaction(value);
      }
    } catch {
      sessionStorage.removeItem(key);
    }
    setReady(true);
  }, []);

  useEffect(() => {
    if (ready && !entry && (!transaction || transaction.step !== mode))
      router.replace(restart);
  }, [ready, entry, transaction, mode, restart, router]);

  useEffect(() => {
    if (mode !== "verify") return;
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, [mode]);

  function save(value: Transaction) {
    sessionStorage.setItem(key, JSON.stringify(value));
    setTransaction(value);
  }

  async function run(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    setNotice("");
    try {
      await action();
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Unable to continue.");
    } finally {
      setBusy(false);
    }
  }

  async function sendCode(value: Transaction) {
    await authRequest("/otp/send", { transaction_id: value.id });
    save({ ...value, step: "verify", resendAt: Date.now() + 60_000 });
    router.push("/email-verification");
  }

  async function submit() {
    if (entry) {
      const result = await authRequest<{ transaction_id: string }>("/start", {
        email: email.trim(),
      });
      const value: Transaction = {
        id: result.transaction_id,
        email: email.trim(),
        purpose:
          mode === "signup"
            ? "signup"
            : mode === "recovery"
              ? "recovery"
              : "login",
        step: "login-password",
        resendAt: 0,
      };
      save(value);
      if (mode === "login") router.push("/log-in/password");
      else await sendCode(value);
      return;
    }
    if (!transaction) return;
    if (mode === "login-password") {
      await authRequest("/password/login", {
        transaction_id: transaction.id,
        password,
      });
    } else if (mode === "verify") {
      await authRequest("/otp/verify", {
        transaction_id: transaction.id,
        code,
      });
      if (transaction.purpose !== "login") {
        const step = transaction.purpose === "signup" ? "enroll" : "reset";
        save({ ...transaction, step });
        router.push(
          step === "enroll" ? "/create-account/password" : "/reset-password",
        );
        return;
      }
    } else {
      if (password !== confirmation)
        throw new Error("The passwords do not match.");
      await authRequest("/password/enroll", { password });
    }
    sessionStorage.removeItem(key);
    router.replace("/");
    router.refresh();
  }

  const remaining = Math.max(
    0,
    Math.ceil(((transaction?.resendAt ?? 0) - now) / 1000),
  );
  if (!ready || (!entry && (!transaction || transaction.step !== mode)))
    return <p role="status">Loading authentication…</p>;

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-semibold tracking-tight">
          {titles[mode]}
        </h1>
        <p className="break-words text-sm text-muted-foreground">
          {entry
            ? mode === "recovery"
              ? "Verify your email to choose a new password."
              : "Continue with your email address."
            : transaction?.email}
        </p>
      </div>
      <form
        className="space-y-4"
        onSubmit={(event) => {
          event.preventDefault();
          void run(submit);
        }}
      >
        {entry && (
          <div className="space-y-2">
            <label htmlFor="email">Email address</label>
            <Input
              id="email"
              type="email"
              autoComplete="email"
              required
              value={email}
              disabled={busy}
              onChange={(event) => setEmail(event.target.value)}
            />
          </div>
        )}
        {mode === "verify" && (
          <div className="space-y-2">
            <label htmlFor="code">Verification code</label>
            <Input
              id="code"
              inputMode="numeric"
              autoComplete="one-time-code"
              pattern="[0-9]{6}"
              maxLength={6}
              required
              value={code}
              disabled={busy}
              onChange={(event) => setCode(event.target.value)}
            />
          </div>
        )}
        {(mode === "login-password" || settingPassword) && (
          <div className="space-y-2">
            <label htmlFor="password">
              {settingPassword ? "New password" : "Password"}
            </label>
            <Input
              id="password"
              type="password"
              autoComplete={
                settingPassword ? "new-password" : "current-password"
              }
              required
              value={password}
              disabled={busy}
              onChange={(event) => setPassword(event.target.value)}
            />
          </div>
        )}
        {settingPassword && (
          <div className="space-y-2">
            <label htmlFor="confirmation">Confirm password</label>
            <Input
              id="confirmation"
              type="password"
              autoComplete="new-password"
              required
              value={confirmation}
              disabled={busy}
              onChange={(event) => setConfirmation(event.target.value)}
            />
          </div>
        )}
        {error && (
          <p role="alert" className="text-sm text-destructive">
            {error}
          </p>
        )}
        <Button className="w-full" type="submit" disabled={busy}>
          {busy
            ? "Please wait…"
            : mode === "verify"
              ? "Verify email"
              : settingPassword
                ? "Save password"
                : "Continue"}
        </Button>
      </form>
      {mode === "login-password" && (
        <div className="flex flex-col items-center gap-3">
          <Button
            variant="outline"
            className="w-full"
            disabled={busy}
            onClick={() => transaction && void run(() => sendCode(transaction))}
          >
            Use an email code instead
          </Button>
          <Link className="text-sm underline" href="/forgot-password">
            Forgot password?
          </Link>
        </div>
      )}
      {mode === "verify" && (
        <Button
          variant="outline"
          className="w-full"
          disabled={busy || remaining > 0}
          onClick={() =>
            transaction &&
            void run(async () => {
              await sendCode(transaction);
              setNotice("A new code has been requested.");
            })
          }
        >
          {remaining > 0 ? `Resend code in ${remaining}s` : "Resend code"}
        </Button>
      )}
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
      <div className="text-center text-sm">
        {mode === "login" ? (
          <p>
            New to Leamout?{" "}
            <Link href="/create-account" className="underline">
              Create account
            </Link>
          </p>
        ) : mode === "signup" ? (
          <p>
            Already have an account?{" "}
            <Link href="/log-in" className="underline">
              Log in
            </Link>
          </p>
        ) : (
          <Link href={restart} className="underline">
            Start again or change email
          </Link>
        )}
      </div>
    </div>
  );
}
