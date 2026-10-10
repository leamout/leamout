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
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const organizationSchema = z.object({
  name: z.string().trim().min(1, "Enter an organization name."),
});

export function CreateOrganizationForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof organizationSchema>>({
    resolver: zodResolver(organizationSchema),
    defaultValues: { name: "" },
  });

  function onSubmit() {
    setNotice("This is a preview. Your organization has not been created.");
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
        onChange={() => setNotice("")}
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
          disabled={form.formState.isSubmitting}
        >
          Create organization
        </Button>
        <p role="status" className="text-sm text-muted-foreground">
          {notice}
        </p>
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
