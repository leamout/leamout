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
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { AIIntegration } from "./types";

const schema = z.object({
  name: z.string().trim().min(1, "Enter a credential name.").max(128),
  secret: z.string().trim().min(1, "Enter an API key."),
});

type Values = z.infer<typeof schema>;

export function ProviderCredentialForm({ providerId }: { providerId: string }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", secret: "" },
  });

  const create = useMutation({
    mutationFn: (values: Values) =>
      api<AIIntegration>("/ai-integrations/", {
        method: "POST",
        organizationId,
        body: {
          provider: providerId,
          name: values.name.trim(),
          secret: values.secret.trim(),
        },
      }),
    onSuccess: async () => {
      form.reset();
      await client.invalidateQueries({
        queryKey: ["ai-integrations", organizationId],
      });
    },
    onSettled: () => form.resetField("secret", { defaultValue: "" }),
  });

  return (
    <form
      className="max-w-2xl space-y-5"
      noValidate
      onSubmit={form.handleSubmit((values) => create.mutate(values))}
    >
      <h2 className="text-lg font-semibold">Add organization credential</h2>
      <Controller
        name="name"
        control={form.control}
        render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="credential-name">Credential name</FieldLabel>
            <Input
              {...field}
              id="credential-name"
              aria-invalid={fieldState.invalid}
              aria-describedby={
                fieldState.invalid ? "credential-name-error" : undefined
              }
              maxLength={128}
              autoComplete="off"
              required
            />
            {fieldState.invalid && (
              <FieldError
                id="credential-name-error"
                errors={[fieldState.error]}
              />
            )}
          </Field>
        )}
      />
      <Controller
        name="secret"
        control={form.control}
        render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="credential-secret">API key</FieldLabel>
            <Input
              {...field}
              id="credential-secret"
              type="password"
              autoComplete="new-password"
              aria-invalid={fieldState.invalid}
              aria-describedby={
                fieldState.invalid ? "credential-secret-error" : undefined
              }
              required
            />
            <FieldDescription>
              Secrets are encrypted by the Go backend and cannot be retrieved
              from this Console. Verify the integration after saving.
            </FieldDescription>
            {fieldState.invalid && (
              <FieldError
                id="credential-secret-error"
                errors={[fieldState.error]}
              />
            )}
          </Field>
        )}
      />
      <Button type="submit" disabled={!organizationId || create.isPending}>
        {create.isPending ? "Saving…" : "Save credential"}
      </Button>
      {create.isSuccess && (
        <p role="status" className="text-sm">
          Credential created. Use Verify to test the connection.
        </p>
      )}
      {create.isError && (
        <p role="alert" className="text-sm text-destructive">
          {errorMessage(create.error)}
        </p>
      )}
    </form>
  );
}
