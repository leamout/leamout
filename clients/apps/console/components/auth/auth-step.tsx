"use client";

import { Skeleton } from "@leamout/ui/components/skeleton";
import Link from "next/link";
import { type ReactNode, useEffect, useState } from "react";
import type { AuthFlow } from "@/lib/auth/flow";
import { useAuthFlow } from "./auth-flow-provider";

export function AuthStep({
  stages,
  children,
}: {
  stages: AuthFlow["stage"][];
  children: ReactNode;
}) {
  const { flow, ready } = useAuthFlow();
  const [now, setNow] = useState(0);
  useEffect(() => {
    setNow(Date.now());
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, []);

  if (!ready || !now) return <Skeleton className="h-48 w-full" />;
  if (!flow || flow.expiresAt <= now || !stages.includes(flow.stage)) {
    return (
      <div className="space-y-4 text-center">
        <p>
          This authentication step is missing or has expired. Start again to
          continue.
        </p>
        <Link
          className="underline"
          href={
            flow?.intent === "recovery"
              ? "/forgot-password"
              : flow?.intent === "create-account"
                ? "/create-account"
                : "/log-in"
          }
        >
          Start again
        </Link>
      </div>
    );
  }
  return children;
}
