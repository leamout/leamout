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
import { errorMessage } from "@/lib/api";
import { authApi } from "@/lib/auth/api";
import { useAuthFlow } from "./auth-flow-provider";
import { AuthForm } from "./auth-form";
import { AuthStep } from "./auth-step";
import { RequireSession } from "./require-session";
import { newPasswordSchema } from "./schemas";

export function CreateAccountPasswordForm() {
  const router = useRouter();
  const client = useQueryClient();
  const { flow, save } = useAuthFlow();
  const mutation = useMutation({ mutationFn: authApi.enroll });
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof newPasswordSchema>>({
    resolver: zodResolver(newPasswordSchema),
    defaultValues: { password: "", confirmation: "" },
  });

  async function onSubmit(values: z.infer<typeof newPasswordSchema>) {
    if (!flow) return;
    setNotice("");
    try {
      await mutation.mutateAsync(values.password);
      await client.cancelQueries();
      client.clear();
      form.reset();
      save(null);
      router.replace("/");
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  return (
    <AuthStep stages={["enroll"]}>
      <RequireSession>
        <AuthForm
          title="Create a password"
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
                    <FieldLabel htmlFor="create-account-password-password">
                      New password
                    </FieldLabel>
                    <Input
                      {...field}
                      id="create-account-password-password"
                      type="password"
                      autoComplete="new-password"
                      required
                      aria-invalid={fieldState.invalid}
                      aria-describedby={
                        fieldState.invalid
                          ? "create-account-password-password-error"
                          : undefined
                      }
                    />
                    {fieldState.invalid && (
                      <FieldError
                        id="create-account-password-password-error"
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
                    <FieldLabel htmlFor="create-account-password-confirmation">
                      Confirm password
                    </FieldLabel>
                    <Input
                      {...field}
                      id="create-account-password-confirmation"
                      type="password"
                      autoComplete="new-password"
                      required
                      aria-invalid={fieldState.invalid}
                      aria-describedby={
                        fieldState.invalid
                          ? "create-account-password-confirmation-error"
                          : undefined
                      }
                    />
                    {fieldState.invalid && (
                      <FieldError
                        id="create-account-password-confirmation-error"
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
      </RequireSession>
    </AuthStep>
  );
}
