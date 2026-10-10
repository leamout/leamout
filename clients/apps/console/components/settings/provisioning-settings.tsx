"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

export type ProvisioningToken = {
  id: string;
  name: string;
  prefix: string;
  expiresAt?: string;
  lastUsedAt?: string;
  revokedAt?: string;
  createdAt: string;
};

const schema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Enter a token name.")
    .max(128, "Use at most 128 characters."),
  expiresAt: z
    .string()
    .refine(
      (value) =>
        !value ||
        (Number.isFinite(Date.parse(value)) && Date.parse(value) > Date.now()),
      "Choose a future expiration time.",
    ),
});

function formatTime(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "—";
  return new Intl.DateTimeFormat("en", {
    dateStyle: "medium",
    timeStyle: "short",
    timeZone: "UTC",
  }).format(date);
}

function tokenStatus(token: ProvisioningToken) {
  if (token.revokedAt) return "Revoked";
  if (token.expiresAt) {
    const expiresAt = Date.parse(token.expiresAt);
    if (!Number.isFinite(expiresAt)) return "Unavailable";
    if (expiresAt <= Date.now()) return "Expired";
  }
  return "Active";
}

function CreateTokenForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", expiresAt: "" },
  });

  return (
    <form
      className="max-w-md space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice("Creation preview only. No token or secret was generated."),
      )}
    >
      {(["name", "expiresAt"] as const).map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`scim-${name}`}>
                {name === "name" ? "Token name" : "Expiration (optional)"}
              </FieldLabel>
              <Input
                {...field}
                id={`scim-${name}`}
                type={name === "expiresAt" ? "datetime-local" : "text"}
                placeholder={
                  name === "name" ? "Directory provisioning" : undefined
                }
                aria-invalid={fieldState.invalid}
                aria-describedby={
                  fieldState.invalid
                    ? `scim-${name}-error`
                    : name === "expiresAt"
                      ? "scim-expiration-help"
                      : undefined
                }
              />
              {name === "expiresAt" && (
                <p
                  id="scim-expiration-help"
                  className="text-xs text-muted-foreground"
                >
                  Uses your local timezone. Leave blank for no expiration.
                </p>
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`scim-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
      ))}
      <Button type="submit">Preview token creation</Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}

export function ProvisioningSettings({
  tokens = [],
  available,
}: {
  tokens?: ProvisioningToken[];
  available?: boolean;
}) {
  const [search, setSearch] = useState("");
  const [revoking, setRevoking] = useState<ProvisioningToken>();
  const [notice, setNotice] = useState("");
  const filtered = tokens.filter((token) =>
    `${token.name} ${token.prefix} ${token.id}`
      .toLowerCase()
      .includes(search.trim().toLowerCase()),
  );

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Provisioning</h1>
        <p className="text-muted-foreground">
          Manage SCIM tokens for directory provisioning.
        </p>
      </div>
      {available === false ? (
        <p className="rounded-lg border p-6 text-sm text-muted-foreground">
          SCIM provisioning requires the SCIM capability for your organization.
        </p>
      ) : (
        <>
          <div className="space-y-2">
            <label htmlFor="scim-search" className="text-sm font-medium">
              Search tokens
            </label>
            <Input
              id="scim-search"
              type="search"
              placeholder="Name, prefix, or token ID"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>
          <div className="overflow-hidden rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  {[
                    "Name",
                    "Prefix",
                    "Status",
                    "Created (UTC)",
                    "Expires (UTC)",
                    "Last used (UTC)",
                    "Actions",
                  ].map((label) => (
                    <TableHead key={label}>{label}</TableHead>
                  ))}
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((token) => (
                  <TableRow key={token.id}>
                    <TableCell className="font-medium">{token.name}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {token.prefix}
                    </TableCell>
                    <TableCell>{tokenStatus(token)}</TableCell>
                    <TableCell>{formatTime(token.createdAt)}</TableCell>
                    <TableCell>
                      {token.expiresAt
                        ? formatTime(token.expiresAt)
                        : "No expiration"}
                    </TableCell>
                    <TableCell>{formatTime(token.lastUsedAt)}</TableCell>
                    <TableCell>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={Boolean(token.revokedAt)}
                        aria-label={`Revoke ${token.name}`}
                        onClick={() => {
                          setRevoking(token);
                          setNotice("");
                        }}
                      >
                        Revoke
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                {filtered.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} className="h-32 text-center">
                      <p className="font-medium">
                        {tokens.length === 0
                          ? "No provisioning tokens to display"
                          : "No matching tokens"}
                      </p>
                      <p className="mt-2 text-muted-foreground">
                        {tokens.length === 0
                          ? "Configured SCIM tokens will appear here."
                          : "Try another search."}
                      </p>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
          {revoking && (
            <div className="space-y-3 rounded-lg border p-6">
              <h2 className="font-semibold">Revoke {revoking.name}?</h2>
              <p className="text-sm text-muted-foreground">
                Revocation stops clients from authenticating with this token.
              </p>
              <div className="flex flex-wrap gap-3">
                <Button
                  variant="destructive"
                  onClick={() => {
                    setRevoking(undefined);
                    setNotice("Revocation preview only. No token was revoked.");
                  }}
                >
                  Confirm revocation preview
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setRevoking(undefined)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          )}
          <p role="status" className="text-sm text-muted-foreground">
            {notice}
          </p>
          <section
            aria-labelledby="create-scim-token-title"
            className="space-y-6 rounded-lg border p-6"
          >
            <div className="space-y-2">
              <h2 id="create-scim-token-title" className="font-semibold">
                Create provisioning token
              </h2>
              <p className="text-sm text-muted-foreground">
                Give each directory integration its own named token. When
                connected, a new token's secret is returned only at creation.
              </p>
            </div>
            <CreateTokenForm />
          </section>
        </>
      )}
      <p className="text-sm text-muted-foreground">
        UI preview. SCIM tokens and organization capabilities are not connected.
      </p>
    </div>
  );
}
