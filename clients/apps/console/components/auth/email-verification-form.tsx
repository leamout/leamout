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
import { verificationSchema } from "./schemas";

export function EmailVerificationForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof verificationSchema>>({
    resolver: zodResolver(verificationSchema),
    defaultValues: { code: "" },
  });

  function onSubmit() {
    setNotice(
      "This form is a preview. Authentication will be connected later.",
    );
  }

  return (
    <AuthForm
      title="Check your email"
      description="Enter the six-digit verification code."
      footer={
        <>
          <p className="text-muted-foreground">
            Resending codes will be available once authentication is connected.
          </p>
          <Link href="/log-in" className="underline">
            Change email
          </Link>
        </>
      }
    >
      <form
        className="space-y-4"
        onSubmit={form.handleSubmit(onSubmit)}
        onChange={() => setNotice("")}
      >
        <FieldGroup>
          <Controller
            name="code"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="email-verification-code">
                  Verification code
                </FieldLabel>
                <Input
                  {...field}
                  id="email-verification-code"
                  type="text"
                  autoComplete="one-time-code"
                  required
                  inputMode="numeric"
                  maxLength={6}
                  pattern="[0-9]{6}"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? "email-verification-code-error"
                      : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id="email-verification-code-error"
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
          Verify email
        </Button>
        <p role="status" className="text-sm text-muted-foreground">
          {notice}
        </p>
      </form>
    </AuthForm>
  );
}
