"use client";

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Button } from "@leamout/ui/components/button";
import { api, errorMessage } from "@/lib/api";

type Session = {
  id: string;
  created_at: string;
  last_seen_at?: string;
  user_agent?: string;
};

export function SecuritySettings() {
  const client = useQueryClient();
  const sessions = useQuery({
    queryKey: ["identity", "sessions"],
    queryFn: () => api<{ sessions: Session[] }>("/sessions/"),
  });
  const revoke = useMutation({
    mutationFn: (id: string) =>
      api<{ message: string }>(`/sessions/${id}`, { method: "DELETE" }),
    onSuccess: () =>
      client.invalidateQueries({ queryKey: ["identity", "sessions"] }),
  });
  const revokeAll = useMutation({
    mutationFn: () =>
      api<{ message: string }>("/sessions/", { method: "DELETE" }),
    onSuccess: () => {
      window.location.assign("/log-in");
    },
  });

  return (
    <div className="space-y-8">
      <section className="space-y-3">
        <h1 className="text-2xl font-semibold">Security</h1>
        <p className="text-muted-foreground">
          Manage your Kinde sign-in profile and Leamout browser sessions.
        </p>
        <a
          href="/api/v1/auth/account"
          className="inline-flex h-9 items-center rounded-md border px-4 text-sm font-medium hover:bg-accent"
        >
          Open Kinde account settings
        </a>
      </section>
      <section className="space-y-4">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <div>
            <h2 className="text-lg font-semibold">Leamout sessions</h2>
            <p className="text-sm text-muted-foreground">
              Revoke browser sessions signed in to Leamout.
            </p>
          </div>
          <Button
            variant="outline"
            disabled={revokeAll.isPending || sessions.isLoading}
            onClick={() => revokeAll.mutate()}
          >
            Sign out all sessions
          </Button>
        </div>
        {sessions.isLoading && (
          <p className="text-sm text-muted-foreground">Loading sessions…</p>
        )}
        {sessions.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(sessions.error)}
          </p>
        )}
        {sessions.data?.sessions.length === 0 && (
          <p className="text-sm text-muted-foreground">No active sessions.</p>
        )}
        <div className="divide-y rounded-lg border">
          {sessions.data?.sessions.map((session) => (
            <div
              key={session.id}
              className="flex items-center justify-between gap-4 p-4"
            >
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">
                  {session.user_agent || "Browser session"}
                </p>
                <p className="text-xs text-muted-foreground">
                  Created {new Date(session.created_at).toLocaleString()}
                </p>
              </div>
              <Button
                size="sm"
                variant="outline"
                disabled={revoke.isPending}
                onClick={() => revoke.mutate(session.id)}
              >
                Revoke
              </Button>
            </div>
          ))}
        </div>
        {(revoke.isError || revokeAll.isError) && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(revoke.error ?? revokeAll.error)}
          </p>
        )}
      </section>
    </div>
  );
}
