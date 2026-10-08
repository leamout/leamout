"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import Link from "next/link";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import type { z } from "zod";
import { AuthForm } from "./auth-form";
import { newPasswordSchema } from "./schemas";

export function ResetPasswordForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof newPasswordSchema>>({
    resolver: zodResolver(newPasswordSchema),
    defaultValues: { password: "", confirmation: "" },
  });

  function onSubmit() {
    setNotice(
      "This form is a preview. Authentication will be connected later.",
    );
  }

  return (
    <AuthForm
      title="Set a new password"
      description="Choose and confirm your password."
      footer={
        <Link href="/log-in" className="underline">
          Back to log in
        </Link>
      }
    >
      <form
        className="space-y-4"
        onSubmit={form.handleSubmit(onSubmit)}
        onChange={() => setNotice("")}
      >
        <FieldGroup>
          <Controller
            name="password"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="reset-password-password">
                  New password
                </FieldLabel>
                <Input
                  {...field}
                  id="reset-password-password"
                  type="password"
                  autoComplete="new-password"
                  required
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? "reset-password-password-error"
                      : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id="reset-password-password-error"
                    errors={[fieldState.error]}
                  />
                )}
              </Field>
            )}
          />
          <Controller
            name="confirmation"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="reset-password-confirmation">
                  Confirm password
                </FieldLabel>
                <Input
                  {...field}
                  id="reset-password-confirmation"
                  type="password"
                  autoComplete="new-password"
                  required
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? "reset-password-confirmation-error"
                      : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id="reset-password-confirmation-error"
                    errors={[fieldState.error]}
                  />
                )}
              </Field>
            )}
          />
        </FieldGroup>
        <Button
          className="w-full"
          type="submit"
          disabled={form.formState.isSubmitting}
        >
          Save password
        </Button>
        <p role="status" className="text-sm text-muted-foreground">
          {notice}
        </p>
      </form>
    </AuthForm>
  );
}
