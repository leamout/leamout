"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { Textarea } from "@leamout/ui/components/textarea";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { errorMessage } from "@/lib/api";
import {
  createVoiceAgent,
  updateVoiceAgent,
  voiceAgentKeys,
} from "./api";
import type { VoiceAgent, VoiceAgentInput } from "./model";

function isJSONObject(value: string): boolean {
  try {
    const parsed: unknown = JSON.parse(value);
    return parsed !== null && !Array.isArray(parsed) && typeof parsed === "object";
  } catch {
    return false;
  }
}

const schema = z.object({
  name: z.string().trim().min(1, "Enter an agent name.").max(255),
  engine: z.enum(["realtime", "composable"]),
  instructions: z
    .string()
    .trim()
    .min(1, "Enter instructions for this agent.")
    .max(20000, "Instructions must be 20,000 characters or fewer."),
  language: z.string().trim().min(1, "Enter a language.").max(64),
  voice: z.string().trim().max(255),
  engineConfig: z.string().refine(isJSONObject, "Enter a valid JSON object."),
  interruptionPolicy: z.enum(["allow", "disabled"]),
  recordingPolicy: z.enum(["none", "all"]),
});

type FormValues = z.infer<typeof schema>;

function defaults(agent?: VoiceAgent): FormValues {
  return {
    name: agent?.name ?? "",
    engine: agent?.engine ?? "realtime",
    instructions: agent?.instructions ?? "",
    voice: agent?.voice ?? "",
    language: agent?.language ?? "en",
    engineConfig: JSON.stringify(agent?.engine_config ?? {}, null, 2),
    interruptionPolicy: agent?.interruption_policy ?? "allow",
    recordingPolicy: agent?.recording_policy ?? "none",
  };
}

function toRequest(values: FormValues, existing?: VoiceAgent): VoiceAgentInput {
  const input: VoiceAgentInput = {
    name: values.name.trim(),
    engine: values.engine,
    instructions: values.instructions.trim(),
    language: values.language.trim(),
    engine_config: JSON.parse(values.engineConfig) as Record<string, unknown>,
    interruption_policy: values.interruptionPolicy,
    recording_policy: values.recordingPolicy,
  };
  // The API cannot clear an existing non-null voice; omitting it preserves
  // the current value. Never claim an empty input removed a saved voice.
  if (values.voice.trim()) {
    input.voice = values.voice.trim();
  } else if (existing?.voice) {
    input.voice = existing.voice;
  }
  return input;
}

export function AgentForm({
  creating = false,
  agent,
}: {
  creating?: boolean;
  agent?: VoiceAgent;
}) {
  const router = useRouter();
  const client = useQueryClient();
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [notice, setNotice] = useState("");
  const createKey = useRef<string | null>(null);
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: defaults(agent),
  });

  const save = useMutation({
    mutationFn: async (values: FormValues) => {
      if (!organizationId) {
        throw new Error("Select an organization before saving.");
      }
      const payload = toRequest(values, agent);
      if (creating) {
        createKey.current ??= crypto.randomUUID();
        return createVoiceAgent(organizationId, payload, createKey.current);
      }
      if (!agent) {
        throw new Error("The agent has not loaded.");
      }
      return updateVoiceAgent(organizationId, agent.id, payload);
    },
    onSuccess: async (saved) => {
      if (!organizationId) return;
      client.setQueryData(voiceAgentKeys.detail(organizationId, saved.id), saved);
      await client.invalidateQueries({
        queryKey: voiceAgentKeys.list(organizationId),
      });
      if (creating) {
        router.push("/agents/" + encodeURIComponent(saved.id));
        return;
      }
      form.reset(defaults(saved));
      setNotice("Changes saved.");
    },
  });

  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => {
        setNotice("");
        if (creating) createKey.current = null;
      }}
      onSubmit={form.handleSubmit((values) => save.mutate(values))}
    >
      <Controller
        name="name"
        control={form.control}
        render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="agent-name">Name</FieldLabel>
            <Input {...field} id="agent-name" required maxLength={255} />
            {fieldState.invalid && (
              <FieldError errors={[fieldState.error]} />
            )}
          </Field>
        )}
      />
      <Field>
        <FieldLabel htmlFor="agent-engine">Conversation engine</FieldLabel>
        <select
          id="agent-engine"
          {...form.register("engine")}
          className="h-10 w-full rounded-md border bg-background px-3 text-sm"
        >
          <option value="realtime">Realtime</option>
          <option value="composable">STT → LLM → TTS</option>
        </select>
      </Field>
      <Controller
        name="instructions"
        control={form.control}
        render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="agent-instructions">Instructions</FieldLabel>
            <Textarea
              {...field}
              id="agent-instructions"
              rows={7}
              required
              maxLength={20000}
              placeholder="Describe the agent's role, behavior, and constraints."
            />
            {fieldState.invalid && (
              <FieldError errors={[fieldState.error]} />
            )}
          </Field>
        )}
      />
      <div className="grid gap-4 sm:grid-cols-2">
        <Controller
          name="language"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="agent-language">Language</FieldLabel>
              <Input {...field} id="agent-language" required maxLength={64} />
              {fieldState.invalid && (
                <FieldError errors={[fieldState.error]} />
              )}
            </Field>
          )}
        />
        <Controller
          name="voice"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="agent-voice">Voice (optional)</FieldLabel>
              <Input {...field} id="agent-voice" maxLength={255} />
              {fieldState.invalid && (
                <FieldError errors={[fieldState.error]} />
              )}
            </Field>
          )}
        />
      </div>
      {agent?.voice && (
        <p className="text-xs text-muted-foreground">
          Clearing a saved voice is not supported by the current API. An empty
          field preserves its existing value.
        </p>
      )}
      <div className="grid gap-4 sm:grid-cols-2">
        <Field>
          <FieldLabel htmlFor="agent-interruption">Interruptions</FieldLabel>
          <select
            {...form.register("interruptionPolicy")}
            id="agent-interruption"
            className="h-10 w-full rounded-md border bg-background px-3 text-sm"
          >
            <option value="allow">Allow</option>
            <option value="disabled">Disabled</option>
          </select>
        </Field>
        <Field>
          <FieldLabel htmlFor="agent-recording">Recording</FieldLabel>
          <select
            {...form.register("recordingPolicy")}
            id="agent-recording"
            className="h-10 w-full rounded-md border bg-background px-3 text-sm"
          >
            <option value="none">Do not record</option>
            <option value="all">Record calls</option>
          </select>
        </Field>
      </div>
      <Controller
        name="engineConfig"
        control={form.control}
        render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="agent-engine-config">
              Advanced engine configuration (JSON)
            </FieldLabel>
            <Textarea
              {...field}
              id="agent-engine-config"
              rows={5}
              spellCheck={false}
              className="font-mono"
            />
            {fieldState.invalid && (
              <FieldError errors={[fieldState.error]} />
            )}
          </Field>
        )}
      />
      <div className="flex flex-wrap gap-3">
        <Button type="submit" disabled={save.isPending || !organizationId}>
          {save.isPending
            ? "Saving…"
            : creating
              ? "Create agent"
              : "Save changes"}
        </Button>
        {creating && (
          <Button variant="outline" render={<Link href="/agents" />}>
            Cancel
          </Button>
        )}
      </div>
      {save.isError && (
        <p role="alert" className="text-sm text-destructive">
          {errorMessage(save.error)}
        </p>
      )}
      {notice && (
        <p role="status" className="text-sm text-muted-foreground">
          {notice}
        </p>
      )}
    </form>
  );
}
