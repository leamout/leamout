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
import { loginPasswordSchema } from "./schemas";

export function LoginPasswordForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof loginPasswordSchema>>({
    resolver: zodResolver(loginPasswordSchema),
    defaultValues: { password: "" },
  });

  function onSubmit() {
    setNotice(
      "This form is a preview. Authentication will be connected later.",
    );
  }

  return (
    <AuthForm
      title="Enter your password"
      description="Use your password to continue."
      footer={
        <>
          <p>
            <Link href="/email-verification" className="underline">
              Use an email code instead
            </Link>
          </p>
          <p>
            <Link href="/forgot-password" className="underline">
              Forgot password?
            </Link>
          </p>
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
            name="password"
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor="login-password-password">
                  Password
                </FieldLabel>
                <Input
                  {...field}
                  id="login-password-password"
                  type="password"
                  autoComplete="current-password"
                  required
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? "login-password-password-error"
                      : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id="login-password-password-error"
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
