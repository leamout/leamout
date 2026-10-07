"use client";

import { useEffect, useState } from "react";

type Envelope = {
  success: boolean;
  data?: { transaction_id?: string };
  error?: { message: string };
};

export default function AcceptInvitationPage() {
  const [token, setToken] = useState("");
  const [email, setEmail] = useState("");
  const [code, setCode] = useState("");
  const [transactionID, setTransactionID] = useState("");
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [accepted, setAccepted] = useState(false);

  useEffect(() => {
    const value =
      new URLSearchParams(window.location.search).get("token") ?? "";
    setToken((current) => value || current);
    // Keep the secret out of subsequent navigation URLs and referrers.
    window.history.replaceState(null, "", window.location.pathname);
  }, []);

  async function request(path: string, body: object): Promise<Envelope> {
    const hostname = window.location.hostname;
    const origin =
      hostname === "localhost" || hostname === "127.0.0.1"
        ? "http://localhost:8080"
        : `https://api.${hostname}`;
    const response = await fetch(`${origin}/v1${path}`, {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    const result = (await response.json()) as Envelope;
    if (!response.ok || !result.success) {
      throw new Error(
        result.error?.message ?? "Request failed. Please try again.",
      );
    }
    return result;
  }

  async function run(action: () => Promise<void>) {
    setBusy(true);
    setMessage("");
    try {
      await action();
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Request failed.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <main className="mx-auto flex min-h-screen max-w-lg flex-col justify-center gap-4 px-6">
      <h1 className="text-2xl font-semibold">Accept your Leamout invitation</h1>
      <p>Sign in with the email address that received the invitation.</p>
      {!accepted && (
        <>
          <button
            type="button"
            disabled={busy || !token}
            onClick={() =>
              run(async () => {
                await request("/invitations/accept", { token });
                setAccepted(true);
                setMessage(
                  "Invitation accepted. Your organization membership is active.",
                );
              })
            }
          >
            Accept using my current session
          </button>
          <form
            className="flex flex-col gap-3"
            onSubmit={(event) => {
              event.preventDefault();
              void run(async () => {
                if (!transactionID) {
                  const result = await request("/auth/start", { email });
                  const id = result.data?.transaction_id;
                  if (!id) throw new Error("Unable to start sign-in.");
                  await request("/auth/otp/send", { transaction_id: id });
                  setTransactionID(id);
                  setMessage("Check your email for a sign-in code.");
                } else {
                  await request("/auth/otp/verify", {
                    transaction_id: transactionID,
                    code,
                  });
                  await request("/invitations/accept", { token });
                  setAccepted(true);
                  setMessage(
                    "Invitation accepted. Your organization membership is active.",
                  );
                }
              });
            }}
          >
            <label htmlFor="invitation-email">Email address</label>
            <input
              id="invitation-email"
              className="rounded border p-2"
              type="email"
              required
              value={email}
              disabled={!!transactionID || busy}
              onChange={(event) => setEmail(event.target.value)}
            />
            {transactionID && (
              <>
                <label htmlFor="invitation-code">Sign-in code</label>
                <input
                  id="invitation-code"
                  className="rounded border p-2"
                  inputMode="numeric"
                  autoComplete="one-time-code"
                  pattern="[0-9]{6}"
                  required
                  value={code}
                  disabled={busy}
                  onChange={(event) => setCode(event.target.value)}
                />
              </>
            )}
            <button type="submit" disabled={busy || !token}>
              {transactionID
                ? "Verify and accept invitation"
                : "Send sign-in code"}
            </button>
          </form>
        </>
      )}
      {!token && (
        <p>
          This invitation link is missing its token. Open the original email
          link.
        </p>
      )}
      <p aria-live="polite">{message}</p>
    </main>
  );
}
