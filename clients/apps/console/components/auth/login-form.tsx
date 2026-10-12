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
import { errorMessage } from "@/lib/api";
import { authApi } from "@/lib/auth/api";
import { useAuthFlow } from "./auth-flow-provider";
import { AuthForm } from "./auth-form";
import { emailSchema } from "./schemas";

export function LoginForm() {
  const router = useRouter();
  const { save, ready } = useAuthFlow();
  const start = useMutation({ mutationFn: authApi.start });
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
        intent: "log-in",
        stage: "password",
        expiresAt: Date.now() + 10 * 60_000,
        resendAt: 0,
      });
      router.push("/log-in/password");
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  return (
    <AuthForm
      title="Log in to Leamout"
      description="Continue with your email address."
      footer={
        <p>
          New to Leamout?{" "}
          <Link href="/create-account" className="underline">
            Create account
          </Link>
        </p>
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
                <FieldLabel htmlFor="login-email">Email address</FieldLabel>
                <Input
                  {...field}
                  id="login-email"
                  type="email"
                  autoComplete="email"
                  required
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? "login-email-error" : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id="login-email-error"
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
