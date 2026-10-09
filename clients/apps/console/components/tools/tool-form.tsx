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
import { Textarea } from "@leamout/ui/components/textarea";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import type { Tool } from "@/components/tools/types";

const builtins = [
  "hangup_call",
  "transfer_call",
  "hold_call",
  "resume_call",
  "send_dtmf",
];
const schema = z
  .object({
    agentId: z.string().min(1, "Choose an agent."),
    type: z.enum(["builtin", "webhook"]),
    name: z.string().trim().min(1, "Enter a tool name.").max(128),
    description: z
      .string()
      .trim()
      .min(1, "Describe when the agent should use this tool.")
      .max(2000),
    parameters: z.string().refine((value) => {
      try {
        const parsed = JSON.parse(value);
        return (
          parsed !== null &&
          typeof parsed === "object" &&
          !Array.isArray(parsed)
        );
      } catch {
        return false;
      }
    }, "Enter a valid JSON object."),
    endpointUrl: z.string(),
    timeoutMs: z
      .string()
      .refine(
        (value) =>
          /^\d+$/.test(value) && Number(value) >= 100 && Number(value) <= 30000,
        "Enter a timeout between 100 and 30000 milliseconds.",
      ),
    enabled: z.boolean(),
  })
  .superRefine((data, context) => {
    if (data.type === "builtin" && !builtins.includes(data.name))
      context.addIssue({
        code: "custom",
        path: ["name"],
        message: "Choose a supported built-in tool.",
      });
    if (data.type === "webhook") {
      try {
        const url = new URL(data.endpointUrl.trim());
        if (url.protocol !== "https:" || !url.hostname) throw new Error();
      } catch {
        context.addIssue({
          code: "custom",
          path: ["endpointUrl"],
          message: "Enter a valid HTTPS endpoint URL.",
        });
      }
    }
  });

export function ToolForm({
  tool,
  creating = false,
  agents = [],
}: {
  tool?: Tool;
  creating?: boolean;
  agents?: { id: string; name: string }[];
}) {
  const [notice, setNotice] = useState("");
  const options =
    tool && !agents.some((agent) => agent.id === tool.agent.id)
      ? [tool.agent, ...agents]
      : agents;
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      agentId: tool?.agent.id ?? "",
      type: tool?.type ?? "webhook",
      name: tool?.name ?? "",
      description: tool?.description ?? "",
      parameters:
        tool?.parameters ??
        JSON.stringify({ type: "object", properties: {} }, null, 2),
      endpointUrl: tool?.endpointUrl ?? "",
      timeoutMs: String(tool?.timeoutMs ?? 10000),
      enabled: tool?.enabled ?? false,
    },
  });
  const type = form.watch("type");
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
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice(
          creating
            ? "Preview only. No tool was created."
            : "Preview only. Changes were not saved.",
        ),
      )}
    >
      {fields.map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`tool-${name}`}>{labels[name]}</FieldLabel>
              {name === "agentId" ? (
                <select
                  {...field}
                  id={`tool-${name}`}
                  disabled={!creating}
                  className="h-9 rounded-md border bg-background px-3"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `tool-${name}-error` : undefined
                  }
                >
                  <option value="">Choose an agent</option>
                  {options.map((agent) => (
                    <option key={agent.id} value={agent.id}>
                      {agent.name}
                    </option>
                  ))}
                </select>
              ) : name === "type" ? (
                <select
                  {...field}
                  id={`tool-${name}`}
                  disabled={!creating}
                  className="h-9 rounded-md border bg-background px-3"
                >
                  <option value="webhook">Webhook</option>
                  <option value="builtin">Built-in call action</option>
                </select>
              ) : name === "name" && type === "builtin" ? (
                <select
                  {...field}
                  id={`tool-${name}`}
                  className="h-9 rounded-md border bg-background px-3"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `tool-${name}-error` : undefined
                  }
                >
                  <option value="">Choose a call action</option>
                  {builtins.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              ) : name === "description" || name === "parameters" ? (
                <Textarea
                  {...field}
                  id={`tool-${name}`}
                  rows={name === "parameters" ? 8 : 3}
                  className={name === "parameters" ? "font-mono" : undefined}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `tool-${name}-error` : undefined
                  }
                />
              ) : (
                <Input
                  {...field}
                  id={`tool-${name}`}
                  type={
                    name === "timeoutMs"
                      ? "number"
                      : name === "endpointUrl"
                        ? "url"
                        : "text"
                  }
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `tool-${name}-error` : undefined
                  }
                />
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`tool-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
              {name === "agentId" && (
                <FieldDescription>
                  Each tool belongs to one agent.{" "}
                  {options.length === 0
                    ? "Agent options have not been loaded."
                    : ""}
                </FieldDescription>
              )}
              {name === "parameters" && (
                <FieldDescription>
                  Validation checks JSON syntax and object shape; full JSON
                  Schema validation is deferred.
                </FieldDescription>
              )}
            </Field>
          )}
        />
      ))}
      <Controller
        name="enabled"
        control={form.control}
        render={({ field }) => (
          <label className="flex items-center gap-3 text-sm font-medium">
            <input
              type="checkbox"
              name={field.name}
              ref={field.ref}
              onBlur={field.onBlur}
              checked={field.value}
              onChange={(event) => field.onChange(event.target.checked)}
            />
            Enable tool
          </label>
        )}
      />
      <Button type="submit" disabled={!creating && !tool}>
        {creating ? "Create tool" : "Save changes"}
      </Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
