"use client";

import { Button } from "@leamout/ui/components/button";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { type FormEvent, useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import type {
  AgentNumberBinding,
  PhoneNumber,
} from "@/components/phone-numbers/types";
import { api, errorMessage } from "@/lib/api";

export function AgentRouting({ agentId }: { agentId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const [phoneNumberId, setPhoneNumberId] = useState("");
  const [removing, setRemoving] = useState<string | null>(null);
  const path = "/voice-agents/" + encodeURIComponent(agentId) + "/bindings";
  const bindingKey = ["agent-number-bindings", organizationId, agentId];

  const bindings = useQuery({
    queryKey: bindingKey,
    queryFn: ({ signal }) =>
      api<{ bindings: AgentNumberBinding[] }>(path, {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const numbers = useQuery({
    queryKey: ["numbers", organizationId],
    queryFn: ({ signal }) =>
      api<{ numbers: PhoneNumber[] }>("/numbers/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const create = useMutation({
    mutationFn: () =>
      api<AgentNumberBinding>(path, {
        method: "POST",
        organizationId,
        body: { phone_number_id: phoneNumberId },
      }),
    onSuccess: async () => {
      setPhoneNumberId("");
      await client.invalidateQueries({ queryKey: bindingKey });
    },
  });

  const remove = useMutation({
    mutationFn: (id: string) =>
      api<void>(path + "/" + encodeURIComponent(id), {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      setRemoving(null);
      await client.invalidateQueries({ queryKey: bindingKey });
    },
  });

  const assigned = new Set(
    bindings.data?.bindings.map((binding) => binding.phone_number_id) ?? [],
  );
  const available = (numbers.data?.numbers ?? []).filter(
    (number) =>
      number.voice_enabled &&
      number.status === "active" &&
      !assigned.has(number.id),
  );

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    if (phoneNumberId) {
      create.mutate();
    }
  }

  return (
    <section className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold">Agent phone number bindings</h2>
        <p className="mt-2 text-sm text-muted-foreground">
          Route incoming calls to this agent through registered BYOC numbers.
        </p>
      </div>

      {bindings.isPending ? (
        <p role="status">Loading number bindings…</p>
      ) : bindings.isError ? (
        <div role="alert">
          <p>{errorMessage(bindings.error)}</p>
          <Button variant="outline" onClick={() => void bindings.refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <div className="space-y-3">
          {bindings.data?.bindings.length ? (
            bindings.data.bindings.map((binding) => {
              const number = numbers.data?.numbers.find(
                (item) => item.id === binding.phone_number_id,
              );

              return (
                <div
                  key={binding.id}
                  className="flex flex-wrap items-center justify-between gap-3 rounded-lg border p-4"
                >
                  <Link
                    href={
                      "/phone-numbers/" +
                      encodeURIComponent(binding.phone_number_id)
                    }
                    className="font-medium hover:underline"
                  >
                    {number?.number ?? binding.phone_number_id}
                  </Link>
                  {removing === binding.id ? (
                    <div className="flex gap-2">
                      <Button
                        variant="destructive"
                        disabled={remove.isPending}
                        onClick={() => remove.mutate(binding.id)}
                      >
                        Confirm removal
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
                      onClick={() => setRemoving(binding.id)}
                    >
                      Unbind
                    </Button>
                  )}
                </div>
              );
            })
          ) : (
            <p className="rounded-lg border border-dashed p-5 text-sm text-muted-foreground">
              No phone numbers assigned to this agent.
            </p>
          )}
        </div>
      )}

      <form
        onSubmit={submit}
        className="max-w-xl space-y-3 rounded-lg border p-5"
      >
        <label
          htmlFor="binding-number"
          className="block text-sm font-medium"
        >
          Assign existing phone number
        </label>
        <select
          id="binding-number"
          className="h-10 w-full rounded-md border bg-background px-3"
          value={phoneNumberId}
          onChange={(event) => setPhoneNumberId(event.target.value)}
          disabled={
            bindings.isPending ||
            numbers.isPending ||
            bindings.isError ||
            numbers.isError
          }
        >
          <option value="">Choose a number</option>
          {available.map((item) => (
            <option key={item.id} value={item.id}>
              {item.number}
            </option>
          ))}
        </select>
        <p className="text-xs text-muted-foreground">
          A number can be bound to only one agent. The server enforces this
          constraint.
        </p>
        {numbers.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(numbers.error)}
          </p>
        )}
        <Button
          type="submit"
          disabled={!phoneNumberId || create.isPending || bindings.isError}
        >
          {create.isPending ? "Assigning…" : "Assign number"}
        </Button>
        {create.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(create.error)}
          </p>
        )}
      </form>
      {remove.isError && (
        <p role="alert" className="text-sm text-destructive">
          {errorMessage(remove.error)}
        </p>
      )}
    </section>
  );
}
