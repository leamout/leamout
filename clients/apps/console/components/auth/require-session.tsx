"use client";

import { Button } from "@leamout/ui/components/button";
import { Skeleton } from "@leamout/ui/components/skeleton";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { type ReactNode, useEffect } from "react";
import { ApiError, api, errorMessage } from "@/lib/api";

export function useSession() {
  return useQuery({
    queryKey: ["session"],
    queryFn: async ({ signal }) => {
      try {
        return await api<{
          id: string;
          email: string;
          name?: string;
          email_verified: boolean;
        }>("/users/me", { signal });
      } catch (error) {
        if (error instanceof ApiError && error.status === 401) return null;
        throw error;
      }
    },
    retry: false,
    staleTime: 0,
  });
}

export function RequireSession({ children }: { children: ReactNode }) {
  const session = useSession();
  const router = useRouter();
  const client = useQueryClient();

  useEffect(() => {
    if (session.data === null) {
      client.removeQueries({
        predicate: (query) => query.queryKey[0] !== "session",
      });
      router.replace("/log-in");
    }
  }, [session.data, router, client]);

  if (session.isError) {
    return (
      <div className="m-auto space-y-4 p-6" role="alert">
        <p>{errorMessage(session.error)}</p>
        <Button onClick={() => session.refetch()}>Try again</Button>
      </div>
    );
  }

  if (!session.data) {
    return (
      <div
        className="mx-auto w-full max-w-sm space-y-4 p-8"
        role="status"
        aria-label="Checking your session"
      >
        <Skeleton className="h-8 w-2/3" />
        <Skeleton className="h-12 w-full" />
      </div>
    );
  }

  return children;
}
