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
import type { WebhookEndpoint } from "@/components/webhooks/types";

const schema = z.object({
  url: z
    .string()
    .trim()
    .refine((value) => {
      try {
        const url = new URL(value);
        return (
          url.protocol === "https:" &&
          Boolean(url.hostname) &&
          !url.username &&
          !url.password
        );
      } catch {
        return false;
      }
    }, "Enter an HTTPS URL without embedded credentials."),
  events: z
    .string()
    .trim()
    .min(1, "Enter at least one subscribed event.")
    .refine(
      (value) =>
        value
          .split("\n")
          .every((event) => event.trim() !== "" && !/\s/.test(event.trim())),
      "Enter one non-empty event name per line.",
    ),
  enabled: z.boolean(),
});

export function WebhookForm({
  endpoint,
  creating = false,
}: {
  endpoint?: WebhookEndpoint;
  creating?: boolean;
}) {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      url: endpoint?.url ?? "",
      events: endpoint?.subscribedEvents.join("\n") ?? "",
      enabled: endpoint?.enabled ?? false,
    },
  });
  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice(
          creating
            ? "Preview only. No webhook endpoint was created."
            : "Preview only. Changes were not saved.",
        ),
      )}
    >
      {(["url", "events"] as const).map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`webhook-${name}`}>
                {name === "url" ? "Endpoint URL" : "Subscribed events"}
              </FieldLabel>
              {name === "url" ? (
                <Input
                  {...field}
                  id={`webhook-${name}`}
                  type="url"
                  placeholder="https://example.com/webhooks/leamout"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `webhook-${name}-error` : undefined
                  }
                />
              ) : (
                <Textarea
                  {...field}
                  id={`webhook-${name}`}
                  rows={6}
                  className="font-mono"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `webhook-${name}-error` : undefined
                  }
                />
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`webhook-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
              {name === "events" && (
                <FieldDescription>
                  Enter one event name per line. The event catalog will be
                  connected later.
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
            Enable endpoint
          </label>
        )}
      />
      <Button type="submit" disabled={!creating && !endpoint}>
        {creating ? "Create endpoint" : "Save changes"}
      </Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
