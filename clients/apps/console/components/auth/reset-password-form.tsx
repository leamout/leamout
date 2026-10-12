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
import { newPasswordSchema } from "./schemas";

export function ResetPasswordForm() {
  const router = useRouter();
  const client = useQueryClient();
  const { flow, save, recoveryCode } = useAuthFlow();
  const mutation = useMutation({
    mutationFn: (body: {
      transaction_id: string;
      code: string;
      password: string;
    }) => api<void>("/auth/password/reset", { method: "POST", body }),
  });
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof newPasswordSchema>>({
    resolver: zodResolver(newPasswordSchema),
    defaultValues: { password: "", confirmation: "" },
  });

  async function onSubmit(values: z.infer<typeof newPasswordSchema>) {
    if (!flow || !recoveryCode) return;
    setNotice("");
    try {
      await mutation.mutateAsync({
        transaction_id: flow.transactionId,
        code: recoveryCode,
        password: values.password,
      });
      await client.cancelQueries();
      client.clear();
      form.reset();
      save(null);
      router.replace("/log-in");
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  return (
    <AuthStep stages={["reset"]}>
      {!recoveryCode ? (
        <div className="space-y-4 text-center">
          <p>Enter your recovery code again to continue.</p>
          <Link href="/email-verification" className="underline">
            Enter recovery code
          </Link>
        </div>
      ) : (
        <AuthForm
          title="Set a new password"
          description="Choose and confirm your password."
          footer={
            <Link href="/email-verification" className="underline">
              Change recovery code
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
            <p role="alert" className="text-sm text-muted-foreground">
              {notice}
            </p>
          </form>
        </AuthForm>
      )}
    </AuthStep>
  );
}
