"use client";

import { Button } from "@leamout/ui/components/button";
import { errorMessage } from "@/lib/api";

export function RequestError({
  error,
  retry,
}: {
  error: unknown;
  retry: () => void;
}) {
  return (
    <div role="alert" className="space-y-3 rounded-lg border p-6">
      <p className="text-sm text-destructive">{errorMessage(error)}</p>
      <Button type="button" variant="outline" onClick={retry}>
        Try again
      </Button>
    </div>
  );
}

export function RequestLoading({ what }: { what: string }) {
  return (
    <p role="status" className="rounded-lg border p-6 text-sm text-muted-foreground">
      Loading {what}…
    </p>
  );
}
