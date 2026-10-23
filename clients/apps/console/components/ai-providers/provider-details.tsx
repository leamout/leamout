"use client";

import { Button } from "@leamout/ui/components/button";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import { providerCatalog, providerName } from "./catalog";
import { CredentialActions } from "./credential-actions";
import { ProviderCredentialForm } from "./provider-credential-form";
import type { AIIntegration } from "./types";

export function ProviderDetails({ providerId }: { providerId: string }) {
  const provider = providerCatalog.find((item) => item.id === providerId);
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const integrations = useQuery({
    queryKey: ["ai-integrations", organizationId],
    queryFn: ({ signal }) =>
      api<{ ai_integrations: AIIntegration[] }>("/ai-integrations/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId) && Boolean(provider),
  });

  if (!provider) {
    return (
      <div className="space-y-4">
        <h1 className="text-3xl font-semibold">Unknown provider</h1>
        <p className="text-muted-foreground">
          This provider is not in the Console's built-in catalog.
        </p>
        <Link href="/ai-providers" className="underline">
          Back to AI providers
        </Link>
      </div>
    );
  }

  const credentials = integrations.data?.ai_integrations.filter(
    (item) => item.provider === providerId,
  ) ?? [];

  return (
    <div className="space-y-8">
      <Link
        href="/ai-providers"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to AI providers
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          {providerName(providerId)}
        </h1>
        <p className="text-muted-foreground">
          Organization-owned AI integrations and their verification state.
        </p>
      </div>

      <section className="space-y-2 rounded-lg border p-6">
        <h2 className="font-semibold">Supported roles</h2>
        <p className="text-sm text-muted-foreground">
          {provider.roles.join(", ").toUpperCase()}
        </p>
        <p className="text-xs text-muted-foreground">
          Provider configuration and role compatibility are checked by the Go
          backend. Platform-owned credentials are not exposed by this API.
        </p>
      </section>

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Organization credentials</h2>
        {integrations.isPending ? (
          <p role="status">Loading credentials…</p>
        ) : integrations.isError ? (
          <div role="alert" className="space-y-3">
            <p>{errorMessage(integrations.error)}</p>
            <Button
              variant="outline"
              onClick={() => void integrations.refetch()}
            >
              Try again
            </Button>
          </div>
        ) : credentials.length ? (
          <ul className="space-y-3">
            {credentials.map((credential) => (
              <li key={credential.id} className="space-y-3 rounded-lg border p-5">
                <div className="space-y-1">
                  <p className="font-medium">{credential.name}</p>
                  <p className="text-sm text-muted-foreground">
                    Connection: {credential.connection_state}
                    {credential.failure_code
                      ? " · " + credential.failure_code
                      : ""}
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Verified:{" "}
                    {credential.verified_at
                      ? new Date(credential.verified_at).toLocaleString()
                      : "Never"}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    Assigned Agents: {credential.voice_agent_ids.length}
                  </p>
                </div>
                <CredentialActions credential={credential} />
              </li>
            ))}
          </ul>
        ) : (
          <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
            No organization credentials configured for this provider.
          </p>
        )}
      </section>

      <ProviderCredentialForm providerId={provider.id} />
      <p className="text-sm text-muted-foreground">
        Verify credentials before assigning them to an{" "}
        <Link href="/agents" className="underline">
          Agent
        </Link>
        .
      </p>
    </div>
  );
}
