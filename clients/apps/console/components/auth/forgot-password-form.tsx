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
import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import type { z } from "zod";
import { api, errorMessage } from "@/lib/api";
import { useAuthFlow } from "./auth-flow-provider";
import { AuthForm } from "./auth-form";
import { emailSchema } from "./schemas";

export function ForgotPasswordForm() {
  const router = useRouter();
  const { save, ready } = useAuthFlow();
  const start = useMutation({
    mutationFn: (email: string) =>
      api<{ transaction_id: string }>("/auth/password/recovery", {
        method: "POST",
        body: { email },
      }),
  });
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof emailSchema>>({
    resolver: zodResolver(emailSchema),
    defaultValues: { email: "" },
  });

  async function onSubmit(values: z.infer<typeof emailSchema>) {
    setNotice("");
    try {
      const transaction = await start.mutateAsync(values.email);
      save({
        transactionId: transaction.transaction_id,
        email: values.email,
        intent: "recovery",
        stage: "verification",
        expiresAt: Date.now() + 10 * 60_000,
        resendAt: Date.now() + 60_000,
      });
      router.push("/email-verification");
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  return (
    <AuthForm
      title="Forgot your password?"
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
                <FieldLabel htmlFor="forgot-password-email">
                  Email address
                </FieldLabel>
                <Input
                  {...field}
                  id="forgot-password-email"
                  type="email"
                  autoComplete="email"
                  required
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? "forgot-password-email-error"
                      : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id="forgot-password-email-error"
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
          disabled={!ready || form.formState.isSubmitting}
        >
          Continue
        </Button>
        <p role="alert" className="text-sm text-muted-foreground">
          {notice}
        </p>
      </form>
    </AuthForm>
  );
}
