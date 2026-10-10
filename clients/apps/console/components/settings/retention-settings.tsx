"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const resources = [
  {
    value: "recordings",
    title: "Recordings",
    description: "Recorded call audio.",
  },
  {
    value: "conversations",
    title: "Conversations",
    description: "Persisted agent conversation data.",
  },
  {
    value: "audit_events",
    title: "Audit events",
    description: "Organization administrative activity.",
  },
] as const;

type Resource = (typeof resources)[number]["value"];

export type RetentionPolicy = {
  resource: Resource;
  retentionDays: number;
  enabled: boolean;
};

const schema = z.object({
  retentionDays: z
    .string()
    .trim()
    .regex(/^\d+$/, "Enter a whole number of days.")
    .refine(
      (value) => Number(value) >= 1 && Number(value) <= 3650,
      "Use between 1 and 3650 days.",
    ),
  enabled: z.boolean(),
});

function ResourcePolicy({
  resource,
  policy,
}: {
  resource: (typeof resources)[number];
  policy?: RetentionPolicy;
}) {
  const [notice, setNotice] = useState("");
  const [confirmRemoval, setConfirmRemoval] = useState(false);
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      retentionDays: policy ? String(policy.retentionDays) : "",
      enabled: policy?.enabled ?? false,
    },
  });

  return (
    <section
      aria-labelledby={`retention-${resource.value}-title`}
      className="space-y-6 rounded-lg border p-6"
    >
      <div className="space-y-2">
        <h2 id={`retention-${resource.value}-title`} className="font-semibold">
          {resource.title}
        </h2>
        <p className="text-sm text-muted-foreground">{resource.description}</p>
        <p className="text-sm">
          {policy
            ? `${policy.enabled ? "Enabled" : "Disabled"} · ${policy.retentionDays} days`
            : "Current policy has not been loaded."}
        </p>
      </div>
      <form
        className="max-w-md space-y-4"
        noValidate
        onChange={() => setNotice("")}
        onSubmit={form.handleSubmit(() =>
          setNotice(
            "Preview only. No retention policy was saved and no data was removed.",
          ),
        )}
      >
        <Controller
          name="retentionDays"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`retention-${resource.value}-days`}>
                Retention days
              </FieldLabel>
              <Input
                {...field}
                id={`retention-${resource.value}-days`}
                inputMode="numeric"
                placeholder="Enter days"
                aria-invalid={fieldState.invalid}
                aria-describedby={`retention-${resource.value}-help${fieldState.invalid ? ` retention-${resource.value}-error` : ""}`}
              />
              <p
                id={`retention-${resource.value}-help`}
                className="text-xs text-muted-foreground"
              >
                A whole number from 1 to 3650.
              </p>
              {fieldState.invalid && (
                <FieldError
                  id={`retention-${resource.value}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
        <Controller
          name="enabled"
          control={form.control}
          render={({ field }) => (
            <div className="flex items-center gap-3">
              <input
                ref={field.ref}
                name={field.name}
                onBlur={field.onBlur}
                id={`retention-${resource.value}-enabled`}
                type="checkbox"
                checked={field.value}
                onChange={(event) => field.onChange(event.target.checked)}
                className="size-4"
              />
              <label
                htmlFor={`retention-${resource.value}-enabled`}
                className="text-sm font-medium"
              >
                Enable policy
              </label>
            </div>
          )}
        />
        <Button type="submit">Preview policy changes</Button>
      </form>
      {confirmRemoval ? (
        <div className="space-y-3">
          <p className="text-sm font-medium">
            Remove the {resource.title.toLowerCase()} policy?
          </p>
          <p className="text-sm text-muted-foreground">
            This removes the policy configuration, rather than deleting the
            resource data directly.
          </p>
          <div className="flex flex-wrap gap-3">
            <Button
              variant="destructive"
              onClick={() => {
                setConfirmRemoval(false);
                setNotice("Removal preview only. No policy was removed.");
              }}
            >
              Confirm removal preview
            </Button>
            <Button variant="outline" onClick={() => setConfirmRemoval(false)}>
              Cancel
            </Button>
          </div>
        </div>
      ) : (
        <Button
          variant="outline"
          disabled={!policy}
          onClick={() => {
            setNotice("");
            setConfirmRemoval(true);
          }}
        >
          Preview policy removal
        </Button>
      )}
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </section>
  );
}

export function RetentionSettings({
  policies = [],
  available,
}: {
  policies?: RetentionPolicy[];
  available?: boolean;
}) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Retention</h1>
        <p className="text-muted-foreground">
          Configure retention periods for recordings, conversations, and audit
          events.
        </p>
      </div>
      {available === false ? (
        <p className="rounded-lg border p-6 text-sm text-muted-foreground">
          Retention policies require the retention policies capability for your
          organization.
        </p>
      ) : (
        <div className="space-y-6">
          {resources.map((resource) => (
            <ResourcePolicy
              key={resource.value}
              resource={resource}
              policy={policies.find(
                (policy) => policy.resource === resource.value,
              )}
            />
          ))}
        </div>
      )}
      <p className="text-sm text-muted-foreground">
        UI preview. Policies and organization capabilities are not connected.
      </p>
    </div>
  );
}
