"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { SipTrunk, TrunkSourceIP } from "./types";

const authSchema = (inbound: boolean) =>
  z
    .object({
      method: z.enum(["ip", "none", "digest"]),
      username: z.string().trim().max(255),
      realm: z.string().trim().max(255),
      secret: z.string().max(4096),
    })
    .superRefine((values, context) => {
      if (
        (inbound && values.method === "none") ||
        (!inbound && values.method === "ip")
      ) {
        context.addIssue({
          code: "custom",
          path: ["method"],
          message: "Select a supported authentication method.",
        });
      }

      if (values.method === "digest") {
        if (!values.username) {
          context.addIssue({
            code: "custom",
            path: ["username"],
            message: "Enter a digest username.",
          });
        }
        if (!values.realm || /[\r\n]/.test(values.realm)) {
          context.addIssue({
            code: "custom",
            path: ["realm"],
            message: "Enter a valid digest realm.",
          });
        }
        if (!values.secret) {
          context.addIssue({
            code: "custom",
            path: ["secret"],
            message: "Enter a new digest password.",
          });
        }
      }
    });

type AuthValues = z.infer<ReturnType<typeof authSchema>>;

const sourceIPSchema = z.object({
  cidr: z
    .string()
    .trim()
    .refine(
      (value) => z.union([z.cidrv4(), z.cidrv6()]).safeParse(value).success,
      "Enter a valid IPv4 or IPv6 network prefix.",
    ),
});

type SourceIPValues = z.infer<typeof sourceIPSchema>;

export function TrunkAuthForm({
  inbound,
  trunk,
}: {
  inbound: boolean;
  trunk: SipTrunk;
}) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const trunkPath = `/trunks/${encodeURIComponent(trunk.id)}`;
  const [removing, setRemoving] = useState<string | null>(null);
  const sourcesKey = ["trunk-source-ips", organizationId, trunk.id];
  const prefix = inbound ? "inbound" : "outbound";

  const form = useForm<AuthValues>({
    resolver: zodResolver(authSchema(inbound)),
    defaultValues: {
      method: inbound ? trunk.inbound_auth_method : trunk.outbound_auth_method,
      username:
        (inbound ? trunk.inbound_username : trunk.outbound_username) ?? "",
      realm: (inbound ? trunk.inbound_realm : trunk.outbound_realm) ?? "",
      secret: "",
    },
  });

  const sourceForm = useForm<SourceIPValues>({
    resolver: zodResolver(sourceIPSchema),
    defaultValues: { cidr: "" },
  });

  const method = form.watch("method");

  const sources = useQuery({
    queryKey: sourcesKey,
    queryFn: ({ signal }) =>
      api<{ source_ips: TrunkSourceIP[] }>(`${trunkPath}/source-ips`, {
        organizationId,
        signal,
      }),
    enabled: inbound && Boolean(organizationId),
  });

  const save = useMutation({
    mutationFn: (values: AuthValues) =>
      api<void>(trunkPath + (inbound ? "/inbound-auth" : "/outbound-auth"), {
        method: "PUT",
        organizationId,
        body:
          values.method === "digest"
            ? {
                method: "digest",
                username: values.username,
                realm: values.realm,
                secret: values.secret,
              }
            : { method: values.method },
      }),
    onSuccess: async () => {
      await client.invalidateQueries({
        queryKey: ["trunks", organizationId, trunk.id],
      });
      await client.invalidateQueries({
        queryKey: ["trunks", organizationId],
      });
    },
    onSettled: () => form.setValue("secret", ""),
  });

  const addIP = useMutation({
    mutationFn: (values: SourceIPValues) =>
      api<TrunkSourceIP>(`${trunkPath}/source-ips`, {
        method: "POST",
        organizationId,
        body: { cidr: values.cidr },
      }),
    onSuccess: async () => {
      sourceForm.reset();
      await client.invalidateQueries({ queryKey: sourcesKey });
    },
  });

  const removeIP = useMutation({
    mutationFn: (id: string) =>
      api<void>(`${trunkPath}/source-ips/${encodeURIComponent(id)}`, {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      setRemoving(null);
      await client.invalidateQueries({ queryKey: sourcesKey });
    },
  });

  return (
    <section className="space-y-5">
      <form
        className="max-w-2xl space-y-5"
        onSubmit={form.handleSubmit((values) => save.mutate(values))}
      >
        <h2 className="text-lg font-semibold">
          {inbound ? "Inbound" : "Outbound"} authentication
        </h2>
        <FieldDescription>
          Current method:{" "}
          {inbound ? trunk.inbound_auth_method : trunk.outbound_auth_method}.
          Existing secrets are never returned by the API. Enter new credentials
          to replace them.
        </FieldDescription>

        <Field data-invalid={Boolean(form.formState.errors.method)}>
          <FieldLabel htmlFor={`${prefix}-method`}>Method</FieldLabel>
          <select
            id={`${prefix}-method`}
            className="h-10 w-full rounded-md border bg-background px-3"
            aria-invalid={Boolean(form.formState.errors.method)}
            {...form.register("method")}
          >
            <option value={inbound ? "ip" : "none"}>
              {inbound ? "Source IP allowlist" : "No digest authentication"}
            </option>
            <option value="digest">Digest authentication</option>
          </select>
          {form.formState.errors.method && (
            <FieldError errors={[form.formState.errors.method]} />
          )}
        </Field>

        {method === "digest" && (
          <div className="space-y-4">
            <Controller
              name="username"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={`${prefix}-username`}>
                    Username
                  </FieldLabel>
                  <Input
                    {...field}
                    id={`${prefix}-username`}
                    autoComplete="off"
                    maxLength={255}
                    required
                    aria-invalid={fieldState.invalid}
                    aria-describedby={
                      fieldState.invalid
                        ? `${prefix}-username-error`
                        : undefined
                    }
                  />
                  {fieldState.invalid && (
                    <FieldError
                      id={`${prefix}-username-error`}
                      errors={[fieldState.error]}
                    />
                  )}
                </Field>
              )}
            />
            <Controller
              name="realm"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={`${prefix}-realm`}>Realm</FieldLabel>
                  <Input
                    {...field}
                    id={`${prefix}-realm`}
                    autoComplete="off"
                    maxLength={255}
                    required
                    aria-invalid={fieldState.invalid}
                    aria-describedby={
                      fieldState.invalid ? `${prefix}-realm-error` : undefined
                    }
                  />
                  {fieldState.invalid && (
                    <FieldError
                      id={`${prefix}-realm-error`}
                      errors={[fieldState.error]}
                    />
                  )}
                </Field>
              )}
            />
            <Controller
              name="secret"
              control={form.control}
              render={({ field, fieldState }) => (
                <Field data-invalid={fieldState.invalid}>
                  <FieldLabel htmlFor={`${prefix}-secret`}>
                    New password
                  </FieldLabel>
                  <Input
                    {...field}
                    id={`${prefix}-secret`}
                    type="password"
                    autoComplete="new-password"
                    maxLength={4096}
                    required
                    aria-invalid={fieldState.invalid}
                    aria-describedby={
                      fieldState.invalid ? `${prefix}-secret-error` : undefined
                    }
                  />
                  {fieldState.invalid && (
                    <FieldError
                      id={`${prefix}-secret-error`}
                      errors={[fieldState.error]}
                    />
                  )}
                </Field>
              )}
            />
          </div>
        )}

        <Button type="submit" disabled={!organizationId || save.isPending}>
          {save.isPending ? "Saving…" : "Save authentication"}
        </Button>
        {save.isSuccess && (
          <p role="status" className="text-sm">
            Authentication updated.
          </p>
        )}
        {save.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(save.error)}
          </p>
        )}
      </form>

      {inbound && (
        <div className="max-w-2xl space-y-4 border-t pt-5">
          <h3 className="font-semibold">Allowed source networks</h3>
          <FieldDescription>
            Source IP networks for inbound authentication.
          </FieldDescription>
          {sources.isPending ? (
            <p role="status">Loading source networks…</p>
          ) : sources.isError ? (
            <div role="alert" className="space-y-3">
              <p>{errorMessage(sources.error)}</p>
              <Button variant="outline" onClick={() => void sources.refetch()}>
                Retry
              </Button>
            </div>
          ) : sources.data?.source_ips.length ? (
            <ul className="space-y-2">
              {sources.data.source_ips.map((item) => (
                <li
                  key={item.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-md border p-3"
                >
                  <code>{item.cidr}</code>
                  {removing === item.id ? (
                    <div className="flex gap-2">
                      <Button
                        variant="destructive"
                        disabled={removeIP.isPending}
                        onClick={() => removeIP.mutate(item.id)}
                      >
                        Confirm
                      </Button>
                      <Button
                        variant="outline"
                        onClick={() => setRemoving(null)}
                      >
                        Cancel
                      </Button>
                    </div>
                  ) : (
                    <Button
                      variant="outline"
                      onClick={() => setRemoving(item.id)}
                    >
                      Remove
                    </Button>
                  )}
                </li>
              ))}
            </ul>
          ) : (
            <p className="text-sm text-muted-foreground">
              No source networks configured.
            </p>
          )}

          <form
            className="flex flex-wrap items-end gap-3"
            onSubmit={sourceForm.handleSubmit((values) => addIP.mutate(values))}
          >
            <Controller
              name="cidr"
              control={sourceForm.control}
              render={({ field, fieldState }) => (
                <Field
                  className="min-w-56 flex-1"
                  data-invalid={fieldState.invalid}
                >
                  <FieldLabel htmlFor="inbound-cidr">
                    Network in CIDR notation
                  </FieldLabel>
                  <Input
                    {...field}
                    id="inbound-cidr"
                    placeholder="203.0.113.0/24"
                    required
                    aria-invalid={fieldState.invalid}
                    aria-describedby={
                      fieldState.invalid ? "inbound-cidr-error" : undefined
                    }
                  />
                  {fieldState.invalid && (
                    <FieldError
                      id="inbound-cidr-error"
                      errors={[fieldState.error]}
                    />
                  )}
                </Field>
              )}
            />
            <Button type="submit" disabled={addIP.isPending || !organizationId}>
              {addIP.isPending ? "Adding…" : "Add network"}
            </Button>
          </form>
          {(addIP.isError || removeIP.isError) && (
            <p role="alert" className="text-sm text-destructive">
              {errorMessage(addIP.error ?? removeIP.error)}
            </p>
          )}
        </div>
      )}
    </section>
  );
}
