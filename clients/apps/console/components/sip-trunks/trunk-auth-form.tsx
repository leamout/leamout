"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState, type FormEvent } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { SipTrunk, TrunkSourceIP } from "./types";

export function TrunkAuthForm({ inbound, trunk }: { inbound: boolean; trunk: SipTrunk }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const trunkPath = "/trunks/" + encodeURIComponent(trunk.id);
  const [method, setMethod] = useState<"ip" | "none" | "digest">(
    inbound ? trunk.inbound_auth_method : trunk.outbound_auth_method,
  );
  const [username, setUsername] = useState("");
  const [realm, setRealm] = useState("");
  const [secret, setSecret] = useState("");
  const [cidr, setCidr] = useState("");
  const [removing, setRemoving] = useState<string | null>(null);
  const sourcesKey = ["trunk-source-ips", organizationId, trunk.id];
  const sources = useQuery({
    queryKey: sourcesKey,
    queryFn: ({ signal }) => api<{ source_ips: TrunkSourceIP[] }>(trunkPath + "/source-ips", { organizationId, signal }),
    enabled: inbound && Boolean(organizationId),
  });
  const save = useMutation({
    mutationFn: () => api<void>(trunkPath + (inbound ? "/inbound-auth" : "/outbound-auth"), {
      method: "PUT",
      organizationId,
      body: method === "digest"
        ? { method: "digest", username: username.trim(), realm: realm.trim(), secret }
        : { method },
    }),
    onSuccess: async () => {
      setSecret("");
      await client.invalidateQueries({ queryKey: ["trunks", organizationId, trunk.id] });
      await client.invalidateQueries({ queryKey: ["trunks", organizationId] });
    },
    onSettled: () => setSecret(""),
  });
  const addIP = useMutation({
    mutationFn: () => api<TrunkSourceIP>(trunkPath + "/source-ips", {
      method: "POST", organizationId, body: { cidr: cidr.trim() },
    }),
    onSuccess: async () => { setCidr(""); await client.invalidateQueries({ queryKey: sourcesKey }); },
  });
  const removeIP = useMutation({
    mutationFn: (id: string) => api<void>(trunkPath + "/source-ips/" + encodeURIComponent(id), {
      method: "DELETE", organizationId,
    }),
    onSuccess: async () => { setRemoving(null); await client.invalidateQueries({ queryKey: sourcesKey }); },
  });
  const prefix = inbound ? "inbound" : "outbound";
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (method === "digest" && (!username.trim() || !realm.trim() || !secret)) return;
    save.mutate();
  }
  return (
    <section className="space-y-5">
      <form className="max-w-2xl space-y-4" onSubmit={submit}>
        <h2 className="text-lg font-semibold">{inbound ? "Inbound" : "Outbound"} authentication</h2>
        <p className="text-sm text-muted-foreground">
          Current method: {inbound ? trunk.inbound_auth_method : trunk.outbound_auth_method}.
          Existing secrets are never returned by the API. Enter new credentials to replace them.
        </p>
        <label className="block space-y-2 text-sm">Method
          <select id={prefix + "-method"} className="block h-10 w-full rounded-md border bg-background px-3"
            value={method} onChange={(e) => setMethod(e.target.value as typeof method)}>
            <option value={inbound ? "ip" : "none"}>{inbound ? "Source IP allowlist" : "No digest authentication"}</option>
            <option value="digest">Digest authentication</option>
          </select>
        </label>
        {method === "digest" && (
          <div className="space-y-4">
            <label className="block space-y-2 text-sm">Username
              <Input value={username} onChange={(e) => setUsername(e.target.value)} required maxLength={255} autoComplete="off" />
            </label>
            <label className="block space-y-2 text-sm">Realm
              <Input value={realm} onChange={(e) => setRealm(e.target.value)} required maxLength={255} autoComplete="off" />
            </label>
            <label className="block space-y-2 text-sm">New password
              <Input type="password" value={secret} onChange={(e) => setSecret(e.target.value)} required maxLength={4096} autoComplete="new-password" />
            </label>
          </div>
        )}
        <Button type="submit" disabled={!organizationId || save.isPending}>
          {save.isPending ? "Saving…" : "Save authentication"}
        </Button>
        {save.isSuccess && <p role="status" className="text-sm">Authentication updated.</p>}
        {save.isError && <p role="alert" className="text-sm text-destructive">{errorMessage(save.error)}</p>}
      </form>
      {inbound && (
        <div className="max-w-2xl space-y-4 border-t pt-5">
          <h3 className="font-semibold">Allowed source networks</h3>
          <p className="text-sm text-muted-foreground">Used for inbound source IP authentication.</p>
          {sources.isPending ? <p role="status">Loading source networks…</p> : sources.isError ? (
            <p role="alert">{errorMessage(sources.error)}</p>
          ) : sources.data?.source_ips.length ? (
            <ul className="space-y-2">
              {sources.data.source_ips.map((item) => (
                <li key={item.id} className="flex flex-wrap items-center justify-between gap-3 rounded-md border p-3">
                  <code>{item.cidr}</code>
                  {removing === item.id ? (
                    <div className="flex gap-2">
                      <Button variant="destructive" disabled={removeIP.isPending} onClick={() => removeIP.mutate(item.id)}>Confirm</Button>
                      <Button variant="outline" onClick={() => setRemoving(null)}>Cancel</Button>
                    </div>
                  ) : <Button variant="outline" onClick={() => setRemoving(item.id)}>Remove</Button>}
                </li>
              ))}
            </ul>
          ) : <p className="text-sm text-muted-foreground">No source networks configured.</p>}
          <form className="flex flex-wrap items-end gap-3" onSubmit={(event) => {
            event.preventDefault();
            if (cidr.trim()) addIP.mutate();
          }}>
            <label className="min-w-56 flex-1 space-y-2 text-sm">Network in CIDR notation
              <Input value={cidr} onChange={(e) => setCidr(e.target.value)} required placeholder="203.0.113.0/24" />
            </label>
            <Button disabled={addIP.isPending || !organizationId} type="submit">Add network</Button>
          </form>
          {(addIP.isError || removeIP.isError) && <p role="alert" className="text-sm text-destructive">{errorMessage(addIP.error ?? removeIP.error)}</p>}
        </div>
      )}
    </section>
  );
}
