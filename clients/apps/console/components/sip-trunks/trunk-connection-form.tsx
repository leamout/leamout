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
import type { TrunkEndpoint } from "./types";

const schema = z.object({
  host: z
    .string()
    .trim()
    .min(1, "Enter a public hostname or IP address.")
    .max(253, "Host must be 253 characters or fewer.")
    .refine(
      (value) => !/[\s/]/.test(value) && !value.includes("://"),
      "Enter a hostname or IP without a scheme or path.",
    ),
  port: z.string().refine(
    (value) =>
      /^\d+$/.test(value) &&
      Number(value) >= 1 &&
      Number(value) <= 65535,
    "Enter a port between 1 and 65535.",
  ),
  transport: z.enum(["udp", "tcp", "tls"]),
  direction: z.enum(["inbound", "outbound", "bidirectional"]),
});

type EndpointValues = z.infer<typeof schema>;

const defaultValues: EndpointValues = {
  host: "",
  port: "5060",
  transport: "udp",
  direction: "bidirectional",
};

export function TrunkConnectionForm({ trunkId }: { trunkId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const base = "/trunks/" + encodeURIComponent(trunkId) + "/endpoints";
  const queryKey = ["trunk-endpoints", organizationId, trunkId];
  const [removing, setRemoving] = useState<string | null>(null);

  const form = useForm<EndpointValues>({
    resolver: zodResolver(schema),
    defaultValues,
  });

  const endpoints = useQuery({
    queryKey,
    queryFn: ({ signal }) =>
      api<{ endpoints: TrunkEndpoint[] }>(base, {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const create = useMutation({
    mutationFn: (values: EndpointValues) =>
      api<TrunkEndpoint>(base, {
        method: "POST",
        organizationId,
        body: {
          host: values.host.trim(),
          port: Number(values.port),
          transport: values.transport,
          direction: values.direction,
        },
      }),
    onSuccess: async () => {
      form.reset();
      await client.invalidateQueries({ queryKey });
    },
  });

  const toggle = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api<TrunkEndpoint>(base + "/" + encodeURIComponent(id), {
        method: "PATCH",
        organizationId,
        body: { enabled },
      }),
    onSuccess: () => client.invalidateQueries({ queryKey }),
  });

  const remove = useMutation({
    mutationFn: (id: string) =>
      api<void>(base + "/" + encodeURIComponent(id), {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      setRemoving(null);
      await client.invalidateQueries({ queryKey });
    },
  });

  return (
    <div className="space-y-7">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold">SIP endpoints</h2>
        <p className="text-sm text-muted-foreground">
          Configure public carrier hosts. Health statuses reflect backend
          observations, not connection guarantees.
        </p>
      </div>

      {endpoints.isPending ? (
        <p role="status">Loading endpoints…</p>
      ) : endpoints.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(endpoints.error)}</p>
          <Button variant="outline" onClick={() => void endpoints.refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {endpoints.data?.endpoints.length ? (
            endpoints.data.endpoints.map((item) => (
              <div
                key={item.id}
                className="flex flex-wrap items-center justify-between gap-4 rounded-lg border p-4"
              >
                <div className="space-y-1">
                  <p className="font-medium">
                    {item.host}:{item.port} · {item.transport.toUpperCase()}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {item.direction} · {item.enabled ? "Enabled" : "Disabled"} ·
                    Health: {item.health_status}
                    {item.last_error ? " · " + item.last_error : ""}
                  </p>
                  {item.last_checked_at && (
                    <p className="text-xs text-muted-foreground">
                      Last checked:{" "}
                      {new Date(item.last_checked_at).toLocaleString()}
                    </p>
                  )}
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    disabled={toggle.isPending || remove.isPending}
                    onClick={() =>
                      toggle.mutate({
                        id: item.id,
                        enabled: !item.enabled,
                      })
                    }
                  >
                    {item.enabled ? "Disable" : "Enable"}
                  </Button>
                  {removing === item.id ? (
                    <>
                      <Button
                        variant="destructive"
                        disabled={remove.isPending}
                        onClick={() => remove.mutate(item.id)}
                      >
                        Confirm
                      </Button>
                      <Button
                        variant="outline"
                        onClick={() => setRemoving(null)}
                      >
                        Cancel
                      </Button>
                    </>
                  ) : (
                    <Button
                      variant="outline"
                      onClick={() => setRemoving(item.id)}
                    >
                      Remove
                    </Button>
                  )}
                </div>
              </div>
            ))
          ) : (
            <p className="rounded-lg border border-dashed p-5 text-sm text-muted-foreground">
              No endpoints configured.
            </p>
          )}
        </div>
      )}

      <form
        className="max-w-2xl space-y-5 rounded-lg border p-5"
        onSubmit={form.handleSubmit((values) => create.mutate(values))}
      >
        <h3 className="font-semibold">Add endpoint</h3>

        <div className="grid gap-4 sm:grid-cols-[1fr_8rem]">
          <Controller
            name="host"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="endpoint-host">
                  Hostname or public IP
                </FieldLabel>
                <Input
                  {...field}
                  id="endpoint-host"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? "endpoint-host-error" : undefined
                  }
                  required
                  placeholder="sip.carrier.example"
                />
                {fieldState.invalid && (
                  <FieldError
                    id="endpoint-host-error"
                    errors={[fieldState.error]}
                  />
                )}
              </Field>
            )}
          />
          <Controller
            name="port"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="endpoint-port">Port</FieldLabel>
                <Input
                  {...field}
                  id="endpoint-port"
                  type="number"
                  min={1}
                  max={65535}
                  step={1}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? "endpoint-port-error" : undefined
                  }
                  required
                />
                {fieldState.invalid && (
                  <FieldError
                    id="endpoint-port-error"
                    errors={[fieldState.error]}
                  />
                )}
              </Field>
            )}
          />
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <Field>
            <FieldLabel htmlFor="endpoint-transport">Transport</FieldLabel>
            <select
              id="endpoint-transport"
              className="h-10 w-full rounded-md border bg-background px-3"
              {...form.register("transport")}
            >
              <option value="udp">UDP</option>
              <option value="tcp">TCP</option>
              <option value="tls">TLS</option>
            </select>
          </Field>
          <Field>
            <FieldLabel htmlFor="endpoint-direction">Direction</FieldLabel>
            <select
              id="endpoint-direction"
              className="h-10 w-full rounded-md border bg-background px-3"
              {...form.register("direction")}
            >
              <option value="bidirectional">Bidirectional</option>
              <option value="inbound">Inbound</option>
              <option value="outbound">Outbound</option>
            </select>
          </Field>
        </div>

        <FieldDescription>
          Endpoints must use a public hostname or IP. The backend also validates
          the host before saving.
        </FieldDescription>
        <Button
          type="submit"
          disabled={!organizationId || create.isPending}
        >
          {create.isPending ? "Adding…" : "Add endpoint"}
        </Button>
        {create.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(create.error)}
          </p>
        )}
      </form>
      {(toggle.isError || remove.isError) && (
        <p role="alert" className="text-sm text-destructive">
          {errorMessage(toggle.error ?? remove.error)}
        </p>
      )}
    </div>
  );
}
