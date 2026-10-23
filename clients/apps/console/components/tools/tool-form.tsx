"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { Textarea } from "@leamout/ui/components/textarea";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import type { VoiceAgent } from "@/components/agents/types";
import { api, errorMessage } from "@/lib/api";
import { SecretReveal } from "@/components/webhooks/secret-reveal";
import { builtins, type Tool } from "./types";

const schema = z.object({
  agentId: z.string().min(1, "Choose an Agent."),
  type: z.enum(["builtin", "webhook"]),
  name: z.string().trim().min(1).max(128),
  description: z.string().trim().min(1).max(2000),
  parameters: z.string().refine((value) => {
    try {
      const data: unknown = JSON.parse(value);
      return data !== null && typeof data === "object" && !Array.isArray(data);
    } catch {
      return false;
    }
  }, "Enter a valid JSON object."),
  endpointUrl: z.string(),
  timeoutMs: z.string().refine(
    (value) => /^\d+$/.test(value) && Number(value) >= 100 && Number(value) <= 30000,
    "Enter a timeout between 100 and 30000 milliseconds.",
  ),
  enabled: z.boolean(),
}).superRefine((values, context) => {
  if (values.type === "builtin" && !builtins.some((name) => name === values.name)) {
    context.addIssue({ code: "custom", path: ["name"], message: "Select a supported built-in tool." });
  }
  if (values.type === "webhook") {
    try {
      const url = new URL(values.endpointUrl);
      if (url.protocol !== "https:" || !url.hostname || url.username || url.password) throw new Error();
    } catch {
      context.addIssue({ code: "custom", path: ["endpointUrl"], message: "Enter a valid HTTPS endpoint." });
    }
  }
});

type Values = z.infer<typeof schema>;

export function ToolForm({
  creating = false,
  tool,
  initialAgentId = "",
}: {
  creating?: boolean;
  tool?: Tool;
  initialAgentId?: string;
}) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const agents = useQuery({
    queryKey: ["voice-agents", organizationId],
    queryFn: ({ signal }) => api<{ voice_agents: VoiceAgent[] }>(
      "/voice-agents/", { organizationId, signal },
    ),
    enabled: Boolean(organizationId),
  });
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      agentId: tool?.voice_agent_id ?? initialAgentId,
      type: tool?.type ?? "webhook",
      name: tool?.name ?? "",
      description: tool?.description ?? "",
      parameters: JSON.stringify(tool?.parameters ?? { type: "object", properties: {} }, null, 2),
      endpointUrl: tool?.endpoint_url ?? "",
      timeoutMs: String(tool?.timeout_ms ?? 10000),
      enabled: tool?.enabled ?? false,
    },
  });
  const type = form.watch("type");
  const save = useMutation({
    mutationFn: (values: Values) => {
      const path = "/voice-agents/" + encodeURIComponent(values.agentId) + "/tools/";
      const body = {
        name: values.name.trim(),
        description: values.description.trim(),
        parameters: JSON.parse(values.parameters) as Record<string, unknown>,
        timeout_ms: Number(values.timeoutMs),
        enabled: values.enabled,
        ...(values.type === "webhook" ? { endpoint_url: values.endpointUrl.trim() } : {}),
      };
      return api<Tool>(
        creating ? path : path + encodeURIComponent(tool?.id ?? ""),
        {
          method: creating ? "POST" : "PATCH",
          organizationId,
          body: creating ? { ...body, type: values.type } : body,
        },
      );
    },
    onSuccess: async (saved) => {
      await client.invalidateQueries({ queryKey: ["tools", organizationId] });
      await client.invalidateQueries({ queryKey: ["agent-tools", organizationId, saved.voice_agent_id] });
      if (!creating) {
        form.reset({
          agentId: saved.voice_agent_id,
          type: saved.type,
          name: saved.name,
          description: saved.description,
          parameters: JSON.stringify(saved.parameters, null, 2),
          endpointUrl: saved.endpoint_url ?? "",
          timeoutMs: String(saved.timeout_ms),
          enabled: saved.enabled,
        });
      }
    },
  });
  const labels = {
    agentId: "Agent",
    type: "Tool type",
    name: "Name",
    description: "Description",
    parameters: "Input parameters (JSON)",
    endpointUrl: "HTTPS endpoint URL",
    timeoutMs: "Timeout (milliseconds)",
  };
  const fields = (Object.keys(labels) as (keyof typeof labels)[]).filter(
    (name) => name !== "endpointUrl" || type === "webhook",
  );
  return (
    <div className="space-y-5">
      <form
        className="max-w-2xl space-y-5"
        noValidate
        onSubmit={form.handleSubmit((values) => save.mutate(values))}
      >
        {fields.map((name) => (
          <Controller
            key={name}
            name={name}
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={"tool-" + name}>{labels[name]}</FieldLabel>
                {name === "agentId" ? (
                  <select
                    {...field}
                    id="tool-agentId"
                    disabled={!creating}
                    className="h-10 rounded-md border bg-background px-3"
                    aria-invalid={fieldState.invalid}
                  >
                    <option value="">Choose Agent</option>
                    {agents.data?.voice_agents.map((agent) => (
                      <option key={agent.id} value={agent.id}>{agent.name}</option>
                    ))}
                    {tool && !agents.data?.voice_agents.some((agent) => agent.id === tool.voice_agent_id) && (
                      <option value={tool.voice_agent_id}>{tool.voice_agent_id}</option>
                    )}
                  </select>
                ) : name === "type" ? (
                  <select
                    {...field}
                    id="tool-type"
                    disabled={!creating}
                    className="h-10 rounded-md border bg-background px-3"
                  >
                    <option value="webhook">Webhook</option>
                    <option value="builtin">Built-in</option>
                  </select>
                ) : name === "name" && type === "builtin" ? (
                  <select {...field} id="tool-name" className="h-10 rounded-md border bg-background px-3">
                    <option value="">Choose built-in action</option>
                    {builtins.map((item) => <option key={item} value={item}>{item}</option>)}
                  </select>
                ) : name === "description" || name === "parameters" ? (
                  <Textarea
                    {...field}
                    id={"tool-" + name}
                    rows={name === "parameters" ? 8 : 3}
                    className={name === "parameters" ? "font-mono" : undefined}
                    aria-invalid={fieldState.invalid}
                  />
                ) : (
                  <Input
                    {...field}
                    id={"tool-" + name}
                    type={name === "timeoutMs" ? "number" : name === "endpointUrl" ? "url" : "text"}
                    aria-invalid={fieldState.invalid}
                  />
                )}
                {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
              </Field>
            )}
          />
        ))}
        <Field orientation="horizontal">
          <input id="tool-enabled" type="checkbox" {...form.register("enabled")} />
          <FieldLabel htmlFor="tool-enabled">Enable tool</FieldLabel>
        </Field>
        <Button type="submit" disabled={save.isPending || !organizationId || (creating && !agents.isSuccess)}>
          {save.isPending ? "Saving…" : creating ? "Create tool" : "Save changes"}
        </Button>
        {save.isError && <p role="alert" className="text-destructive">{errorMessage(save.error)}</p>}
        {save.isSuccess && !creating && <p role="status">Tool updated.</p>}
      </form>
      {creating && save.data && (
        <div className="space-y-3">
          {save.data.signing_secret && (
            <SecretReveal secret={save.data.signing_secret} label="Tool signing secret" />
          )}
          <Button
            variant="outline"
            onClick={() => router.push(
              "/tools/" + encodeURIComponent(save.data.id) +
              "?agent_id=" + encodeURIComponent(save.data.voice_agent_id),
            )}
          >
            View tool
          </Button>
        </div>
      )}
    </div>
  );
}
