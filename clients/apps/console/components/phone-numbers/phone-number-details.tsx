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
import { PhoneNumberForm } from "./phone-number-form";
import { PhoneNumberRoutingForm } from "./phone-number-routing-form";
import type { PhoneNumber } from "./types";

export function PhoneNumberDetails({ numberId }: { numberId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const [confirm, setConfirm] = useState(false);
  const path = "/numbers/" + encodeURIComponent(numberId);

  const number = useQuery({
    queryKey: ["numbers", organizationId, numberId],
    queryFn: ({ signal }) => api<PhoneNumber>(path, { organizationId, signal }),
    enabled: Boolean(organizationId),
  });

  const remove = useMutation({
    mutationFn: () =>
      api<void>(path, {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      client.removeQueries({
        queryKey: ["numbers", organizationId, numberId],
        exact: true,
      });
      await client.invalidateQueries({
        queryKey: ["numbers", organizationId],
      });
      router.replace("/phone-numbers");
    },
  });

  if (number.isPending) {
    return <p role="status">Loading phone number…</p>;
  }

  if (number.isError) {
    return (
      <div role="alert" className="space-y-3">
        <p>{errorMessage(number.error)}</p>
        <Button variant="outline" onClick={() => void number.refetch()}>
          Retry
        </Button>
      </div>
    );
  }

  const item = number.data;

  return (
    <div className="space-y-8">
      <Link
        href="/phone-numbers"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to phone numbers
      </Link>
      <div>
        <h1 className="text-3xl font-semibold">{item.number}</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          {item.country_code} · {item.status} · Voice{" "}
          {item.voice_enabled ? "enabled" : "disabled"}
        </p>
      </div>

      <Tabs defaultValue="configuration" className="gap-6">
        <TabsList aria-label="Phone number settings">
          <TabsTrigger value="configuration">Configuration</TabsTrigger>
          <TabsTrigger value="routing">Agent routing</TabsTrigger>
        </TabsList>
        <TabsContent value="configuration">
          <PhoneNumberForm key={item.id} number={item} />
        </TabsContent>
        <TabsContent value="routing">
          <PhoneNumberRoutingForm number={item} />
        </TabsContent>
      </Tabs>

      <section className="space-y-3 border-t pt-8">
        <h2 className="font-semibold">Remove registered number</h2>
        <p className="text-sm text-muted-foreground">
          This does not cancel service with your carrier.
        </p>
        {!confirm ? (
          <Button variant="outline" onClick={() => setConfirm(true)}>
            Remove number…
          </Button>
        ) : (
          <div className="flex gap-2">
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
