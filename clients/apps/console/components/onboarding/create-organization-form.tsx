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
import Link from "next/link";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { api, errorMessage } from "@/lib/api";
import type { Organization } from "@/lib/organizations";
import { useOrganization } from "./organization-provider";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const organizationSchema = z.object({
  name: z.string().trim().min(1, "Enter an organization name."),
});

export function CreateOrganizationForm() {
  const router = useRouter();
  const organization = useOrganization();
  const create = useMutation({
    mutationFn: (name: string) =>
      api<Organization>("/organizations", {
        method: "POST",
        body: { name },
      }),
    onSuccess: (created) => {
      if (organization.addOrganization(created)) {
        router.replace("/");
      }
    },
  });
  const form = useForm<z.infer<typeof organizationSchema>>({
    resolver: zodResolver(organizationSchema),
    defaultValues: { name: "" },
  });

  function onSubmit(values: z.infer<typeof organizationSchema>) {
    create.mutate(values.name.trim());
  }

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-semibold tracking-tight">
          Create an organization
        </h1>
        <p className="text-sm text-muted-foreground">
          A workspace for your team, voice agents, and connections.
        </p>
      </div>
      <form
        className="space-y-4"
        onSubmit={form.handleSubmit(onSubmit)}
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
                autoComplete="organization"
                required
                aria-invalid={fieldState.invalid}
                aria-describedby={
                  fieldState.invalid
                    ? "organization-name-help organization-name-error"
                    : "organization-name-help"
                }
              />
              <FieldDescription id="organization-name-help">
                Use your company or team name.
              </FieldDescription>
              {fieldState.invalid && (
                <FieldError
                  id="organization-name-error"
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
        <Button
          className="w-full"
          type="submit"
          disabled={create.isPending}
        >
          Create organization
        </Button>
        {create.isError && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(create.error)}
          </p>
        )}
        {create.isSuccess && !organization.selected && (
          <p role="alert" className="text-sm text-destructive">
            Organization created, but your browser could not save the selection.
            Select it from the organization list to continue.
          </p>
        )}
      </form>
      <p className="text-center text-sm">
        Already have an organization?{" "}
        <Link href="/organizations" className="underline">
          Select organization
        </Link>
      </p>
    </div>
  );
}
