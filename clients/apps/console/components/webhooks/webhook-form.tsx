"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { Textarea } from "@leamout/ui/components/textarea";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { WebhookEndpoint, WebhookSecretResponse } from "./types";
import { SecretReveal } from "./secret-reveal";

const schema = z.object({
  url: z.url().refine((value) => {
    const parsed = new URL(value);
    return parsed.protocol === "https:" && !parsed.username && !parsed.password;
  }, "Enter an HTTPS URL without credentials."),
  events: z.string().trim().min(1, "Enter at least one event.").refine(
    (value) => value.split("\n").every((name) => name.trim() && !/\s/.test(name.trim())),
    "Use one event type per line.",
  ),
  enabled: z.boolean(),
});

type Values = z.infer<typeof schema>;

export function WebhookForm({
  endpoint,
  creating = false,
}: {
  endpoint?: WebhookEndpoint;
  creating?: boolean;
}) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      url: endpoint?.url ?? "",
      events: endpoint?.subscribed_events.join("\n") ?? "",
      enabled: endpoint?.enabled ?? false,
    },
  });
  const save = useMutation({
    mutationFn: async (values: Values) => {
      const body = {
        url: values.url.trim(),
        subscribed_events: [...new Set(values.events.split("\n").map((item) => item.trim()))],
        enabled: values.enabled,
      };
      if (creating) {
        return api<WebhookSecretResponse>("/webhooks/", {
          method: "POST",
          organizationId,
          body,
        });
      }
      if (!endpoint) throw new Error("Webhook endpoint not loaded.");
      const saved = await api<WebhookEndpoint>(
        "/webhooks/" + encodeURIComponent(endpoint.id),
        { method: "PATCH", organizationId, body },
      );
      return { webhook: saved, signing_secret: "" };
    },
    onSuccess: async (value) => {
      client.setQueryData(["webhooks", organizationId, value.webhook.id], value.webhook);
      await client.invalidateQueries({ queryKey: ["webhooks", organizationId] });
      if (!creating) {
        form.reset({
          url: value.webhook.url,
          events: value.webhook.subscribed_events.join("\n"),
          enabled: value.webhook.enabled,
        });
      }
    },
  });

  return (
    <div className="space-y-5">
      <form
        className="max-w-2xl space-y-5"
        noValidate
        onSubmit={form.handleSubmit((values) => save.mutate(values))}
      >
        <Controller
          name="url"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="webhook-url">Endpoint HTTPS URL</FieldLabel>
              <Input {...field} id="webhook-url" type="url" aria-invalid={fieldState.invalid} />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />
        <Controller
          name="events"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="webhook-events">Subscribed events (one per line)</FieldLabel>
              <Textarea {...field} id="webhook-events" rows={6} className="font-mono" aria-invalid={fieldState.invalid} />
              {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
            </Field>
          )}
        />
        <Field orientation="horizontal">
          <input id="webhook-enabled" type="checkbox" {...form.register("enabled")} />
          <FieldLabel htmlFor="webhook-enabled">Enable endpoint</FieldLabel>
        </Field>
        <Button type="submit" disabled={!organizationId || save.isPending}>
          {save.isPending ? "Saving…" : creating ? "Create endpoint" : "Save changes"}
        </Button>
        {save.isError && (
          <p role="alert" className="text-sm text-destructive">{errorMessage(save.error)}</p>
        )}
        {save.isSuccess && !creating && <p role="status">Endpoint updated.</p>}
      </form>
      {creating && save.data && (
        <div className="space-y-3">
          <SecretReveal secret={save.data.signing_secret} label="Webhook signing secret" />
          <Button variant="outline" onClick={() => router.push("/webhooks/" + encodeURIComponent(save.data.webhook.id))}>
            View endpoint
          </Button>
        </div>
      )}
    </div>
  );
}
