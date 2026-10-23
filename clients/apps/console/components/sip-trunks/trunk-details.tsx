"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import { TrunkAuthForm } from "./trunk-auth-form";
import { TrunkConnectionForm } from "./trunk-connection-form";
import { TrunkForm } from "./trunk-form";
import type { SipTrunk, TrunkValidation } from "./types";

export function TrunkDetails({ trunkId }: { trunkId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const [confirm, setConfirm] = useState(false);
  const path = `/trunks/${encodeURIComponent(trunkId)}`;

  const query = useQuery({
    queryKey: ["trunks", organizationId, trunkId],
    queryFn: ({ signal }) => api<SipTrunk>(path, { organizationId, signal }),
    enabled: Boolean(organizationId),
  });

  const validate = useMutation({
    mutationFn: () =>
      api<TrunkValidation>(`${path}/validate`, {
        method: "POST",
        organizationId,
      }),
  });

  const remove = useMutation({
    mutationFn: () =>
      api<void>(path, {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      client.removeQueries({
        queryKey: ["trunks", organizationId, trunkId],
        exact: true,
      });
      await client.invalidateQueries({
        queryKey: ["trunks", organizationId],
      });
      router.replace("/sip-trunks");
    },
  });

  if (query.isPending) {
    return <p role="status">Loading SIP trunk…</p>;
  }

  if (query.isError) {
    return (
      <div role="alert" className="space-y-3">
        <p>{errorMessage(query.error)}</p>
        <Button variant="outline" onClick={() => void query.refetch()}>
          Retry
        </Button>
      </div>
    );
  }

  const trunk = query.data;

  return (
    <div className="space-y-8">
      <Link
        href="/sip-trunks"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to SIP trunks
      </Link>

      <div className="space-y-2">
        <h1 className="text-3xl font-semibold">{trunk.name}</h1>
        <p className="font-mono text-xs text-muted-foreground">{trunk.id}</p>
        <p className="text-sm text-muted-foreground">
          {trunk.direction} · {trunk.status} ·{" "}
          {trunk.inbound_enabled ? "Inbound enabled" : "Inbound disabled"}
        </p>
      </div>

      <section className="space-y-3 rounded-lg border p-5">
        <h2 className="font-semibold">Configuration validation</h2>
        <Button
          variant="outline"
          disabled={validate.isPending}
          onClick={() => validate.mutate()}
        >
          {validate.isPending ? "Validating…" : "Validate SIP trunk"}
        </Button>
        {validate.isError && (
          <p role="alert" className="text-destructive">
            {errorMessage(validate.error)}
          </p>
        )}
        {validate.data && (
          <div role="status" className="space-y-2 text-sm">
            <p className="font-medium">
              {validate.data.valid
                ? "Configuration valid"
                : "Configuration needs attention"}
            </p>
            {validate.data.errors.map((item) => (
              <p key={item} className="text-destructive">
                {item}
              </p>
            ))}
            {validate.data.warnings.map((item) => (
              <p key={item} className="text-muted-foreground">
                Warning: {item}
              </p>
            ))}
          </div>
        )}
        <p className="text-xs text-muted-foreground">
          Validation checks configuration, not guaranteed carrier connectivity.
        </p>
      </section>

      <Tabs defaultValue="configuration" className="gap-6">
        <div className="overflow-x-auto">
          <TabsList aria-label="SIP trunk settings">
            <TabsTrigger value="configuration">Configuration</TabsTrigger>
            <TabsTrigger value="endpoints">Endpoints</TabsTrigger>
            <TabsTrigger value="authentication">Authentication</TabsTrigger>
            <TabsTrigger value="routing">Numbers</TabsTrigger>
          </TabsList>
        </div>
        <TabsContent value="configuration">
          <TrunkForm key={trunk.id} trunk={trunk} />
        </TabsContent>
        <TabsContent value="endpoints">
          <TrunkConnectionForm trunkId={trunk.id} />
        </TabsContent>
        <TabsContent value="authentication">
          <div className="space-y-10">
            <TrunkAuthForm key={`in-${trunk.id}`} inbound trunk={trunk} />
            <TrunkAuthForm
              key={`out-${trunk.id}`}
              inbound={false}
              trunk={trunk}
            />
          </div>
        </TabsContent>
        <TabsContent value="routing">
          <p className="text-muted-foreground">
            Connect your own numbers and assign them to this trunk from Phone
            Numbers.
          </p>
          <Link href="/phone-numbers" className="text-sm underline">
            Manage phone numbers
          </Link>
        </TabsContent>
      </Tabs>

      <section className="space-y-3 border-t pt-8">
        <h2 className="font-semibold">Remove SIP trunk</h2>
        {!confirm ? (
          <Button variant="outline" onClick={() => setConfirm(true)}>
            Remove trunk…
          </Button>
        ) : (
          <div className="flex flex-wrap gap-2">
            <Button
              variant="destructive"
              disabled={remove.isPending}
              onClick={() => remove.mutate()}
            >
              Confirm removal
            </Button>
            <Button variant="outline" onClick={() => setConfirm(false)}>
              Cancel
            </Button>
          </div>
        )}
        {remove.isError && (
          <p role="alert" className="text-destructive">
            {errorMessage(remove.error)}
          </p>
        )}
      </section>
    </div>
  );
}
