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
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import type { z } from "zod";
import { api, errorMessage } from "@/lib/api";
import { useAuthFlow } from "./auth-flow-provider";
import { AuthForm } from "./auth-form";
import { AuthStep } from "./auth-step";
import { loginPasswordSchema } from "./schemas";

export function LoginPasswordForm() {
  const router = useRouter();
  const client = useQueryClient();
  const { flow, save } = useAuthFlow();
  const login = useMutation({
    mutationFn: (body: { transaction_id: string; password: string }) =>
      api<void>("/auth/password/login", { method: "POST", body }),
  });
  const send = useMutation({
    mutationFn: (transaction_id: string) =>
      api<void>("/auth/otp/send", { method: "POST", body: { transaction_id } }),
  });
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof loginPasswordSchema>>({
    resolver: zodResolver(loginPasswordSchema),
    defaultValues: { password: "" },
  });

  async function onSubmit(values: z.infer<typeof loginPasswordSchema>) {
    if (!flow) return;
    setNotice("");
    try {
      await login.mutateAsync({
        transaction_id: flow.transactionId,
        password: values.password,
      });
      await client.cancelQueries();
      client.clear();
      form.reset();
      save(null);
      router.replace("/");
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  async function sendCode() {
    if (!flow || send.isPending || login.isPending) return;
    setNotice("");
    try {
      await send.mutateAsync(flow.transactionId);
      save({ ...flow, stage: "verification", resendAt: Date.now() + 60_000 });
      router.push("/email-verification");
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  return (
    <AuthStep stages={["password"]}>
      <AuthForm
        title="Enter your password"
        description={`Continue as ${flow?.email ?? ""}.`}
        footer={
          <>
            <p>
              <button
                type="button"
                onClick={sendCode}
                disabled={send.isPending || login.isPending}
                className="underline disabled:opacity-50"
              >
                {send.isPending ? "Sending code…" : "Use an email code instead"}
              </button>
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
            disabled={form.formState.isSubmitting || send.isPending}
          >
            Continue
          </Button>
          <p role="alert" className="text-sm text-muted-foreground">
            {notice}
          </p>
        </form>
      </AuthForm>
    </AuthStep>
  );
}
