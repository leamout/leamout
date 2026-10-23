"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { AIIntegration } from "./types";

const schema = z.object({
  secret: z.string().trim().min(1, "Enter a replacement API key."),
});

type Values = z.infer<typeof schema>;

export function CredentialActions({
  credential,
}: {
  credential: AIIntegration;
}) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const [rotating, setRotating] = useState(false);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { secret: "" },
  });

  const path = "/ai-integrations/" + encodeURIComponent(credential.id);
  const refresh = () =>
    client.invalidateQueries({
      queryKey: ["ai-integrations", organizationId],
    });

  const verify = useMutation({
    mutationFn: () =>
      api<{
        id: string;
        connection_state: string;
        verified_at?: string | null;
        failure_code?: string | null;
      }>(path + "/verify", { method: "POST", organizationId }),
    onSuccess: refresh,
  });

  const rotate = useMutation({
    mutationFn: (values: Values) =>
      api<AIIntegration>(path + "/rotate", {
        method: "POST",
        organizationId,
        body: { secret: values.secret.trim() },
      }),
    onSuccess: async () => {
      setRotating(false);
      await refresh();
    },
    onSettled: () => form.resetField("secret", { defaultValue: "" }),
  });

  const remove = useMutation({
    mutationFn: () =>
      api<void>(path, { method: "DELETE", organizationId }),
    onSuccess: async () => {
      setConfirmDelete(false);
      await refresh();
    },
  });

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap gap-2">
        <Button
          variant="outline"
          disabled={verify.isPending || !organizationId}
          onClick={() => verify.mutate()}
        >
          {verify.isPending ? "Verifying…" : "Verify"}
        </Button>
        <Button
          variant="outline"
          onClick={() => {
            setRotating((value) => !value);
            setConfirmDelete(false);
          }}
        >
          {rotating ? "Cancel rotation" : "Rotate key"}
        </Button>
        <Button
          variant="outline"
          disabled={remove.isPending}
          onClick={() => {
            setConfirmDelete(true);
            setRotating(false);
          }}
        >
          Remove
        </Button>
      </div>
      {verify.isSuccess && (
        <p role="status" className="text-sm">
          Verification completed: {verify.data.connection_state}
          {verify.data.failure_code ? " · " + verify.data.failure_code : ""}
        </p>
      )}
      {verify.isError && (
        <p role="alert" className="text-sm text-destructive">
          {errorMessage(verify.error)}
        </p>
      )}

      {rotating && (
        <form
          className="max-w-md space-y-3 rounded-md border p-4"
          noValidate
          onSubmit={form.handleSubmit((values) => rotate.mutate(values))}
        >
          <Controller
            name="secret"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={"rotate-" + credential.id}>
                  Replacement API key
                </FieldLabel>
                <Input
                  {...field}
                  id={"rotate-" + credential.id}
                  type="password"
                  autoComplete="new-password"
                  aria-invalid={fieldState.invalid}
                  required
                />
                {fieldState.invalid && (
                  <FieldError errors={[fieldState.error]} />
                )}
              </Field>
            )}
          />
          <Button type="submit" disabled={rotate.isPending || !organizationId}>
            {rotate.isPending ? "Rotating…" : "Save replacement"}
          </Button>
          {rotate.isError && (
            <p role="alert" className="text-sm text-destructive">
              {errorMessage(rotate.error)}
            </p>
          )}
          <p className="text-xs text-muted-foreground">
            Rotated credentials must be verified again.
          </p>
        </form>
      )}

      {confirmDelete && (
        <div className="space-y-3 rounded-md border p-4">
          <p className="text-sm">
            Remove this organization credential? Agents using it may become
            unavailable.
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              variant="destructive"
              disabled={remove.isPending}
              onClick={() => remove.mutate()}
            >
              {remove.isPending ? "Removing…" : "Confirm removal"}
            </Button>
            <Button
              variant="outline"
              disabled={remove.isPending}
              onClick={() => setConfirmDelete(false)}
            >
              Cancel
            </Button>
          </div>
          {remove.isError && (
            <p role="alert" className="text-sm text-destructive">
              {errorMessage(remove.error)}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
