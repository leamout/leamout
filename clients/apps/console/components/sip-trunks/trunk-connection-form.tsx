"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { type FormEvent, useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { TrunkEndpoint } from "./types";

export function TrunkConnectionForm({ trunkId }: { trunkId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const base = "/trunks/" + encodeURIComponent(trunkId) + "/endpoints";
  const key = ["trunk-endpoints", organizationId, trunkId];
  const [host, setHost] = useState("");
  const [port, setPort] = useState("5060");
  const [transport, setTransport] = useState<"udp" | "tcp" | "tls">("udp");
  const [direction, setDirection] = useState<
    "inbound" | "outbound" | "bidirectional"
  >("bidirectional");
  const [removing, setRemoving] = useState<string | null>(null);

  const endpoints = useQuery({
    queryKey: key,
    queryFn: ({ signal }) =>
      api<{ endpoints: TrunkEndpoint[] }>(base, {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const create = useMutation({
    mutationFn: () =>
      api<TrunkEndpoint>(base, {
        method: "POST",
        organizationId,
        body: {
          host: host.trim(),
          port: Number(port),
          transport,
          direction,
        },
      }),
    onSuccess: async () => {
      setHost("");
      await client.invalidateQueries({ queryKey: key });
    },
  });

  const toggle = useMutation({
    mutationFn: ({ id, enabled }: { id: string; enabled: boolean }) =>
      api<TrunkEndpoint>(base + "/" + encodeURIComponent(id), {
        method: "PATCH",
        organizationId,
        body: { enabled },
      }),
    onSuccess: () => client.invalidateQueries({ queryKey: key }),
  });

  const remove = useMutation({
    mutationFn: (id: string) =>
      api<void>(base + "/" + encodeURIComponent(id), {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      setRemoving(null);
      await client.invalidateQueries({ queryKey: key });
    },
  });

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (
      host.trim() &&
      Number.isInteger(Number(port)) &&
      Number(port) > 0 &&
      Number(port) <= 65535
    ) {
      create.mutate();
    }
  }

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
        <div role="alert">
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
        className="max-w-2xl space-y-4 rounded-lg border p-5"
        onSubmit={submit}
      >
        <h3 className="font-semibold">Add endpoint</h3>
        <div className="grid gap-4 sm:grid-cols-[1fr_8rem]">
          <label className="space-y-2 text-sm">
            Hostname or public IP
            <Input
              value={host}
              onChange={(event) => setHost(event.target.value)}
              required
              placeholder="sip.carrier.example"
            />
          </label>
          <label className="space-y-2 text-sm">
            Port
            <Input
              value={port}
              onChange={(event) => setPort(event.target.value)}
              required
              type="number"
              min={1}
              max={65535}
              step={1}
            />
          </label>
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <label className="space-y-2 text-sm">
            Transport
            <select
              className="block h-10 w-full rounded-md border bg-background px-3"
              value={transport}
              onChange={(event) =>
                setTransport(event.target.value as typeof transport)
              }
            >
              <option value="udp">UDP</option>
              <option value="tcp">TCP</option>
              <option value="tls">TLS</option>
            </select>
          </label>
          <label className="space-y-2 text-sm">
            Direction
            <select
              className="block h-10 w-full rounded-md border bg-background px-3"
              value={direction}
              onChange={(event) =>
                setDirection(event.target.value as typeof direction)
              }
            >
              <option value="bidirectional">Bidirectional</option>
              <option value="inbound">Inbound</option>
              <option value="outbound">Outbound</option>
            </select>
          </label>
        </div>
        <Button disabled={!organizationId || create.isPending} type="submit">
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
