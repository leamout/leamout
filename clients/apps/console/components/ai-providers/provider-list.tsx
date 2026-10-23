"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import { providerCatalog } from "./catalog";
import type { AIIntegration } from "./types";

export function ProviderList() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [search, setSearch] = useState("");

  const integrations = useQuery({
    queryKey: ["ai-integrations", organizationId],
    queryFn: ({ signal }) =>
      api<{ ai_integrations: AIIntegration[] }>("/ai-integrations/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const filtered = providerCatalog.filter((provider) =>
    provider.name.toLowerCase().includes(search.trim().toLowerCase()),
  );

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">AI providers</h1>
        <p className="text-muted-foreground">
          Create and verify organization-owned AI integrations for Voice Agents.
        </p>
      </div>

      <div className="max-w-md space-y-2">
        <label htmlFor="provider-search" className="text-sm font-medium">
          Search providers
        </label>
        <Input
          id="provider-search"
          type="search"
          value={search}
          onChange={(event) => setSearch(event.target.value)}
        />
      </div>

      {integrations.isPending ? (
        <p role="status">Loading AI integrations…</p>
      ) : integrations.isError ? (
        <div className="space-y-3" role="alert">
          <p>{errorMessage(integrations.error)}</p>
          <Button variant="outline" onClick={() => void integrations.refetch()}>
            Try again
          </Button>
        </div>
      ) : (
        <div className="overflow-hidden rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Provider</TableHead>
                <TableHead>Roles</TableHead>
                <TableHead>Integrations</TableHead>
                <TableHead>Verified</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.map((provider) => {
                const credentials = integrations.data.ai_integrations.filter(
                  (item) => item.provider === provider.id,
                );
                const ready = credentials.filter(
                  (item) => item.connection_state === "ready",
                ).length;

                return (
                  <TableRow key={provider.id}>
                    <TableCell>
                      <Link
                        className="font-medium hover:underline"
                        href={"/ai-providers/" + encodeURIComponent(provider.id)}
                      >
                        {provider.name}
                      </Link>
                    </TableCell>
                    <TableCell>{provider.roles.join(", ").toUpperCase()}</TableCell>
                    <TableCell>{credentials.length}</TableCell>
                    <TableCell>{ready}</TableCell>
                  </TableRow>
                );
              })}
              {filtered.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="py-12 text-center">
                    No matching providers.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      )}
      <p className="text-xs text-muted-foreground">
        Role labels describe the built-in providers. The backend validates role
        compatibility and credential ownership when an Agent is configured.
        Platform credential availability is not exposed by this API.
      </p>
    </div>
  );
}
