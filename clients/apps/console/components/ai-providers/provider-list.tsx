"use client";

import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import Link from "next/link";
import { useState } from "react";
import {
  type ProviderAvailability,
  providerCatalog,
} from "@/components/ai-providers/catalog";

export function ProviderList({
  availability = [],
}: {
  availability?: ProviderAvailability[];
}) {
  const [search, setSearch] = useState("");
  const filtered = providerCatalog.filter((provider) =>
    provider.name.toLowerCase().includes(search.toLowerCase()),
  );
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">AI providers</h1>
        <p className="text-muted-foreground">
          Configure organization credentials and review platform defaults for
          your agents.
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
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {[
                "Provider",
                "Capabilities",
                "Organization BYOA",
                "Platform default",
              ].map((label) => (
                <TableHead key={label}>{label}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((provider) => {
              const state = availability.find(
                (item) => item.providerId === provider.id,
              );
              return (
                <TableRow key={provider.id}>
                  <TableCell>
                    <Link
                      className="font-medium hover:underline"
                      href={`/ai-providers/${provider.id}`}
                    >
                      {provider.name}
                    </Link>
                  </TableCell>
                  <TableCell>
                    {state?.capabilities
                      ? state.capabilities.join(", ") || "None available"
                      : "Not loaded"}
                  </TableCell>
                  <TableCell>
                    {state?.organizationCredentialCount === undefined
                      ? "Not loaded"
                      : `${state.organizationCredentialCount} credentials`}
                  </TableCell>
                  <TableCell>
                    {state?.platformAvailable === undefined
                      ? "Not loaded"
                      : state.platformAvailable
                        ? "Available"
                        : "Unavailable"}
                  </TableCell>
                </TableRow>
              );
            })}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="h-32 text-center">
                  No matching providers. Try another search.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Capability and credential availability have not been
        fetched. Choose credential sources within agent configuration.
      </p>
    </div>
  );
}
