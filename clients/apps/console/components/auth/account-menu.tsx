"use client";

import { Button } from "@leamout/ui/components/button";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { ApiError, errorMessage } from "@/lib/api";
import { authApi } from "@/lib/auth/api";
import { flowKey } from "@/lib/auth/flow";
import { useSession } from "./require-session";

export function AccountMenu() {
  const session = useSession();
  const client = useQueryClient();
  const router = useRouter();
  const logout = useMutation({
    mutationFn: async () => {
      try {
        await authApi.logout();
      } catch (error) {
        if (!(error instanceof ApiError && error.status === 401)) throw error;
      }
    },
    onSuccess: async () => {
      await client.cancelQueries();
      client.clear();
      try {
        sessionStorage.removeItem(flowKey);
      } catch {
        // Session storage may be disabled.
      }
      router.replace("/log-in");
      router.refresh();
    },
  });

  return (
    <div className="space-y-2 px-2 py-2 group-data-[collapsible=icon]:hidden">
      <p className="truncate text-xs text-muted-foreground">
        {session.data?.email}
      </p>
      <Button
        variant="ghost"
        size="sm"
        disabled={logout.isPending}
        onClick={() => logout.mutate()}
      >
        {logout.isPending ? "Logging out…" : "Log out"}
      </Button>
      {logout.isError && (
        <p role="alert" className="text-xs text-destructive">
          {errorMessage(logout.error)}
        </p>
      )}
    </div>
  );
}
