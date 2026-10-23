"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Textarea } from "@leamout/ui/components/textarea";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import type { VoiceAgent } from "@/components/agents/types";
import {
  providerCatalog,
  providerName,
} from "@/components/ai-providers/catalog";
import type {
  AgentProviderBinding,
  AIIntegration,
  ProviderRole,
} from "@/components/ai-providers/types";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";

function isJSONObject(value: string): boolean {
  try {
    const parsed: unknown = JSON.parse(value);
    return (
      parsed !== null && !Array.isArray(parsed) && typeof parsed === "object"
    );
  } catch {
    return false;
  }
}

const schema = z.object({
  provider: z.string().trim().min(1, "Select a provider."),
  credentialId: z.uuid("Select an organization credential."),
  config: z.string().refine(isJSONObject, "Enter a valid JSON object."),
});

type Values = z.infer<typeof schema>;

function ProviderRoleForm({
  role,
  agentId,
  organizationId,
  credentials,
  binding,
}: {
  role: ProviderRole;
  agentId: string;
  organizationId: string;
  credentials: AIIntegration[];
  binding?: AgentProviderBinding;
}) {
  const client = useQueryClient();
  const [confirmRemove, setConfirmRemove] = useState(false);
  const [saved, setSaved] = useState(false);
  const key = ["agent-provider-bindings", organizationId, agentId];
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      provider: binding?.provider ?? "",
      credentialId: binding?.credential_id ?? "",
      config: JSON.stringify(binding?.config ?? {}, null, 2),
    },
  });
  const provider = form.watch("provider");
  const options = providerCatalog.filter((entry) => entry.roles.includes(role));
  const matchingCredentials = credentials.filter(
    (entry) =>
      entry.provider === provider && entry.connection_state === "ready",
  );

  const refresh = async () => {
    await client.invalidateQueries({ queryKey: key });
    await client.invalidateQueries({
      queryKey: ["agent-readiness", organizationId, agentId],
    });
    await client.invalidateQueries({
      queryKey: ["ai-integrations", organizationId],
    });
  };

  const save = useMutation({
    mutationFn: (values: Values) =>
      api<AgentProviderBinding>(
        "/voice-agents/" +
          encodeURIComponent(agentId) +
          "/providers/" +
          encodeURIComponent(role),
        {
          method: "PUT",
          organizationId,
          body: {
            provider: values.provider,
            credential_id: values.credentialId,
            config: JSON.parse(values.config) as Record<string, unknown>,
          },
        },
      ),
    onSuccess: async (value) => {
      form.reset({
        provider: value.provider,
        credentialId: value.credential_id ?? "",
        config: JSON.stringify(value.config ?? {}, null, 2),
      });
      setSaved(true);
      await refresh();
    },
  });

  const remove = useMutation({
    mutationFn: () =>
      api<void>(
        "/voice-agents/" +
          encodeURIComponent(agentId) +
          "/providers/" +
          encodeURIComponent(role),
        { method: "DELETE", organizationId },
      ),
    onSuccess: async () => {
      setConfirmRemove(false);
      setSaved(false);
      form.reset({ provider: "", credentialId: "", config: "{}" });
      await refresh();
    },
  });

  return (
    <section className="space-y-4 rounded-lg border p-6">
      <h3 className="font-semibold uppercase">{role}</h3>
      <p className="text-sm text-muted-foreground">
        {binding
          ? `Current assignment: ${providerName(binding.provider)}`
          : "No provider assigned."}
      </p>
      <form
        className="space-y-4"
        noValidate
        onSubmit={form.handleSubmit((values) => save.mutate(values))}
        onChange={() => setSaved(false)}
      >
        <Field data-invalid={Boolean(form.formState.errors.provider)}>
          <FieldLabel htmlFor={`provider-${role}`}>Provider</FieldLabel>
          <select
            id={`provider-${role}`}
            className="h-10 w-full rounded-md border bg-background px-3"
            aria-invalid={Boolean(form.formState.errors.provider)}
            {...form.register("provider", {
              onChange: () => form.setValue("credentialId", ""),
            })}
          >
            <option value="">Choose provider</option>
            {options.map((option) => (
              <option key={option.id} value={option.id}>
                {option.name}
              </option>
            ))}
          </select>
          {form.formState.errors.provider && (
            <FieldError errors={[form.formState.errors.provider]} />
          )}
        </Field>
        <Field data-invalid={Boolean(form.formState.errors.credentialId)}>
          <FieldLabel htmlFor={`credential-${role}`}>
            Verified organization credential
          </FieldLabel>
          <select
            id={`credential-${role}`}
            className="h-10 w-full rounded-md border bg-background px-3"
            aria-invalid={Boolean(form.formState.errors.credentialId)}
            {...form.register("credentialId")}
          >
            <option value="">Choose credential</option>
            {matchingCredentials.map((credential) => (
              <option key={credential.id} value={credential.id}>
                {credential.name}
              </option>
            ))}
          </select>
          {form.formState.errors.credentialId && (
            <FieldError errors={[form.formState.errors.credentialId]} />
          )}
          {!matchingCredentials.length && (
            <p className="text-sm text-muted-foreground">
              No verified credentials for this provider.{" "}
              <Link href="/ai-providers" className="underline">
                Manage integrations
              </Link>
            </p>
          )}
        </Field>
        <Controller
          name="config"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`config-${role}`}>
                Provider configuration (JSON)
              </FieldLabel>
              <Textarea
                {...field}
                id={`config-${role}`}
                className="font-mono"
                rows={5}
                spellCheck={false}
                aria-invalid={fieldState.invalid}
              />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />
        <p className="text-xs text-muted-foreground">
          Enter provider configuration, not API secrets. The backend validates
          provider-specific settings and credential ownership.
        </p>
        <Button type="submit" disabled={save.isPending || !organizationId}>
          {save.isPending ? "Saving…" : `Save ${role.toUpperCase()}`}
        </Button>
        {saved && (
          <p role="status" className="text-sm">
            Assignment saved.
          </p>
        )}
        {save.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(save.error)}
          </p>
        )}
      </form>

      {binding && (
        <div className="space-y-2 border-t pt-4">
          {!confirmRemove ? (
            <Button variant="outline" onClick={() => setConfirmRemove(true)}>
              Remove assignment…
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
              <Button variant="outline" onClick={() => setConfirmRemove(false)}>
                Cancel
              </Button>
            </div>
          )}
          {remove.isError && (
            <p role="alert" className="text-sm text-destructive">
              {errorMessage(remove.error)}
            </p>
          )}
        </div>
      )}
    </section>
  );
}

export function AgentProviderSettings({ agent }: { agent: VoiceAgent }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const path = `/voice-agents/${encodeURIComponent(agent.id)}/providers/`;
  const roles: ProviderRole[] =
    agent.engine === "realtime" ? ["realtime"] : ["stt", "llm", "tts"];

  const bindings = useQuery({
    queryKey: ["agent-provider-bindings", organizationId, agent.id],
    queryFn: ({ signal }) =>
      api<{ providers: AgentProviderBinding[] }>(path, {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  const integrations = useQuery({
    queryKey: ["ai-integrations", organizationId],
    queryFn: ({ signal }) =>
      api<{ ai_integrations: AIIntegration[] }>("/ai-integrations/", {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold">Provider assignments</h2>
        <p className="text-sm text-muted-foreground">
          {agent.engine === "realtime"
            ? "Configure one realtime provider for this Agent."
            : "Configure STT, LLM, and TTS providers for this Agent."}{" "}
          Each role requires a verified organization integration.
        </p>
        <Link href="/ai-providers" className="text-sm underline">
          Manage provider credentials
        </Link>
      </div>

      {bindings.isPending || integrations.isPending ? (
        <p role="status">Loading provider assignments…</p>
      ) : bindings.isError || integrations.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(bindings.error ?? integrations.error)}</p>
          <Button
            variant="outline"
            onClick={() => {
              void bindings.refetch();
              void integrations.refetch();
            }}
          >
            Try again
          </Button>
        </div>
      ) : (
        <>
          {roles.map((role) => {
            const binding = bindings.data.providers.find(
              (item) => item.role === role,
            );
            return (
              <ProviderRoleForm
                key={
                  agent.id +
                  "-" +
                  role +
                  "-" +
                  (binding?.id ?? "none") +
                  "-" +
                  organizationId
                }
                role={role}
                agentId={agent.id}
                organizationId={organizationId}
                credentials={integrations.data.ai_integrations}
                binding={binding}
              />
            );
          })}
          {bindings.data.providers.some(
            (item) => !roles.includes(item.role),
          ) && (
            <p className="text-sm text-muted-foreground">
              This Agent has bindings for a different engine. Use its
              Configuration tab to review the engine or remove incompatible
              bindings when you switch back.
            </p>
          )}
        </>
      )}
    </div>
  );
}
