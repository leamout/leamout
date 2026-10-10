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
import { emailSchema } from "./schemas";

export function CreateAccountForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof emailSchema>>({
    resolver: zodResolver(emailSchema),
    defaultValues: { email: "" },
  });

  function onSubmit() {
    setNotice(
      "This form is a preview. Authentication will be connected later.",
    );
  }

  return (
    <AuthForm
      title="Create your account"
      description="Continue with your email address."
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
            name="email"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="create-account-email">
                  Email address
                </FieldLabel>
                <Input
                  {...field}
                  id="create-account-email"
                  type="email"
                  autoComplete="email"
                  required
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? "create-account-email-error"
                      : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id="create-account-email-error"
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
          Continue
        </Button>
        <p role="status" className="text-sm text-muted-foreground">
          {notice}
        </p>
      </form>
    </AuthForm>
  );
}
