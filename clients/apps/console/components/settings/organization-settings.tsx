"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import {
  Field,
  FieldError,
  FieldLabel,
} from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { Organization } from "@/lib/organizations";

const schema = z.object({
  name: z.string().trim().min(1, "Enter an organization name.").max(128),
});

type Values = z.infer<typeof schema>;

function OrganizationForm({
  organization,
  canEdit,
}: {
  organization: Organization;
  canEdit: boolean;
}) {
  const client = useQueryClient();

  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: { name: organization.name },
  });

  const save = useMutation({
    mutationFn: (values: Values) =>
      api<Organization>("/organizations/" + encodeURIComponent(organization.id), {
        method: "PATCH",
        organizationId: organization.id,
        body: { name: values.name.trim() },
      }),
    onSuccess: async (value) => {
      form.reset({ name: value.name });
      client.setQueryData(
        ["organizations", value.id],
        value,
      );
      await client.invalidateQueries({ queryKey: ["organizations"] });
    },
  });

  return (
    <form
      className="max-w-lg space-y-5"
      noValidate
      onSubmit={form.handleSubmit((values) => save.mutate(values))}
    >
      <Controller
        name="name"
        control={form.control}
        render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="organization-name">
              Organization name
            </FieldLabel>
            <Input
              {...field}
              id="organization-name"
              maxLength={128}
              required
              disabled={!canEdit}
              aria-invalid={fieldState.invalid}
              aria-describedby={
                fieldState.invalid ? "organization-name-error" : undefined
              }
            />
            {fieldState.invalid && (
              <FieldError
                id="organization-name-error"
                errors={[fieldState.error]}
              />
            )}
          </Field>
        )}
      />
      <Button type="submit" disabled={!canEdit || save.isPending}>
        {save.isPending ? "Saving…" : "Save organization"}
      </Button>
      {save.isSuccess && <p role="status">Organization updated.</p>}
      {save.isError && (
        <p role="alert" className="text-sm text-destructive">
          {errorMessage(save.error)}
        </p>
      )}
    </form>
  );
}

export function OrganizationSettings() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const canEdit = selected?.role === "owner" || selected?.role === "admin";

  const organization = useQuery({
    queryKey: ["organization-details", organizationId],
    queryFn: ({ signal }) =>
      api<Organization>("/organizations/" + encodeURIComponent(organizationId), {
        organizationId,
        signal,
      }),
    enabled: Boolean(organizationId),
  });

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">Organization</h1>
        <p className="text-muted-foreground">
          Review and manage your selected organization.
        </p>
      </div>
      {organization.isPending ? (
        <p role="status">Loading organization…</p>
      ) : organization.isError ? (
        <div role="alert" className="space-y-3">
          <p>{errorMessage(organization.error)}</p>
          <Button variant="outline" onClick={() => void organization.refetch()}>
            Retry
          </Button>
        </div>
      ) : (
        <>
          <dl className="grid gap-4 rounded-lg border p-5 sm:grid-cols-2">
            <div>
              <dt className="text-sm text-muted-foreground">Organization ID</dt>
              <dd className="break-all font-mono text-xs">{organization.data.id}</dd>
            </div>
            <div>
              <dt className="text-sm text-muted-foreground">Your role</dt>
              <dd className="font-medium capitalize">{selected?.role}</dd>
            </div>
            <div>
              <dt className="text-sm text-muted-foreground">Status</dt>
              <dd className="font-medium capitalize">{organization.data.status}</dd>
            </div>
            <div>
              <dt className="text-sm text-muted-foreground">Created</dt>
              <dd>{organization.data.created_at
                ? new Date(organization.data.created_at).toLocaleString()
                : "—"}</dd>
            </div>
          </dl>
          <OrganizationForm
            key={organization.data.id + organization.data.updated_at}
            organization={organization.data}
            canEdit={canEdit}
          />
          {!canEdit && (
            <p className="text-sm text-muted-foreground">
              Only organization owners and administrators can edit details.
            </p>
          )}
        </>
      )}
    </div>
  );
}
