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
import { useEffect, useState } from "react";
import { Controller, useForm } from "react-hook-form";
import type { z } from "zod";
import { errorMessage } from "@/lib/api";
import { authApi } from "@/lib/auth/api";
import { useAuthFlow } from "./auth-flow-provider";
import { AuthForm } from "./auth-form";
import { AuthStep } from "./auth-step";
import { verificationSchema } from "./schemas";

export function EmailVerificationForm() {
  const router = useRouter();
  const client = useQueryClient();
  const { flow, save, setRecoveryCode } = useAuthFlow();
  const verify = useMutation({ mutationFn: authApi.verify });
  const send = useMutation({ mutationFn: authApi.sendCode });
  const [now, setNow] = useState(0);
  useEffect(() => {
    setNow(Date.now());
    const timer = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(timer);
  }, []);
  const wait = Math.max(0, Math.ceil(((flow?.resendAt ?? 0) - now) / 1000));
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof verificationSchema>>({
    resolver: zodResolver(verificationSchema),
    defaultValues: { code: "" },
  });

  async function onSubmit(values: z.infer<typeof verificationSchema>) {
    if (!flow) return;
    setNotice("");
    if (flow.intent === "recovery") {
      setRecoveryCode(values.code);
      save({ ...flow, stage: "reset" });
      form.reset();
      router.push("/reset-password");
      return;
    }
    try {
      await verify.mutateAsync({
        transaction_id: flow.transactionId,
        code: values.code,
      });
      await client.cancelQueries();
      client.clear();
      form.reset();
      if (flow.intent === "create-account") {
        save({ ...flow, stage: "enroll" });
        router.replace("/create-account/password");
      } else {
        save(null);
        router.replace("/");
      }
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  async function resend() {
    if (!flow || wait > 0 || send.isPending || verify.isPending) return;
    setNotice("");
    try {
      await send.mutateAsync(flow.transactionId);
      setRecoveryCode("");
      save({ ...flow, stage: "verification", resendAt: Date.now() + 60_000 });
      setNotice("If this address is eligible, a new code will arrive shortly.");
    } catch (error) {
      setNotice(errorMessage(error));
    }
  }

  return (
    <AuthStep stages={["verification", "reset"]}>
      <AuthForm
        title="Check your email"
        description={
          flow?.intent === "recovery"
            ? `If an account exists for ${flow.email}, we sent a recovery code. Enter it to continue.`
            : `Enter the six-digit code sent to ${flow?.email ?? "your email"}.`
        }
        footer={
          <>
            <button
              type="button"
              className="underline disabled:opacity-50"
              onClick={resend}
              disabled={wait > 0 || send.isPending || verify.isPending}
            >
              {send.isPending
                ? "Sending…"
                : wait > 0
                  ? `Resend code in ${wait}s`
                  : "Resend code"}
            </button>
            <Link
              href={
                flow?.intent === "recovery"
                  ? "/forgot-password"
                  : flow?.intent === "create-account"
                    ? "/create-account"
                    : "/log-in"
              }
              className="underline"
            >
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
            disabled={form.formState.isSubmitting || send.isPending}
          >
            {flow?.intent === "recovery" ? "Continue" : "Verify email"}
          </Button>
          <p role="status" className="text-sm text-muted-foreground">
            {notice}
          </p>
        </form>
      </AuthForm>
    </AuthStep>
  );
}
