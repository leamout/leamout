import { RequireSession } from "@/components/auth/require-session";
import { ConsoleShell } from "@/components/console-shell";
import { OrganizationProvider } from "@/components/onboarding/organization-provider";
import { RequireOrganization } from "@/components/onboarding/require-organization";

export default function ConsoleLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <RequireSession>
      <OrganizationProvider>
        <RequireOrganization>
          <ConsoleShell>{children}</ConsoleShell>
        </RequireOrganization>
      </OrganizationProvider>
    </RequireSession>
  );
}
