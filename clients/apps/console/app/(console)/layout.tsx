import { RequireSession } from "@/components/auth/require-session";
import { ConsoleShell } from "@/components/console-shell";

export default function ConsoleLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <RequireSession>
      <ConsoleShell>{children}</ConsoleShell>
    </RequireSession>
  );
}
