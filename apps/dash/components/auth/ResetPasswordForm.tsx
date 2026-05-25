"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";

const formSchema = z
  .object({
    password: z.string().min(8, "Password must be at least 8 characters."),
    confirmPassword: z.string().min(1, "Please confirm your password."),
  })
  .refine((values) => values.password === values.confirmPassword, {
    message: "Passwords do not match.",
    path: ["confirmPassword"],
  });

type FormValues = z.infer<typeof formSchema>;

export function ResetPasswordForm() {
  const [loading, setLoading] = useState(false);
  const searchParams = useSearchParams();
  const token = searchParams.get("token");

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: { password: "", confirmPassword: "" },
  });

  const onSubmit = async ({ confirmPassword: _, password }: FormValues) => {
    if (!token) {
      toast.error("Missing reset token.");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("/api/v1/auth/reset-password", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ token, password }),
      });

      if (response.ok) {
        toast.success("Password reset successful. You can now login.");
        form.reset();
        return;
      }

      const errorData = await response.json();
      toast.error(errorData.message || "Failed to reset password.");
    } catch {
      toast.error("An unexpected error occurred.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
      <FieldGroup className="space-y-2">
        {(["password", "confirmPassword"] as const).map((name) => (
          <Controller
            key={name}
            name={name}
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={name}>
                  {name === "confirmPassword" ? "Confirm password" : "Password"}
                </FieldLabel>
                <Input
                  {...field}
                  id={name}
                  type="password"
                  autoComplete="new-password"
                  disabled={loading || !token}
                />
                {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
              </Field>
            )}
          />
        ))}
      </FieldGroup>
      <div className="space-y-4">
        <Button type="submit" disabled={loading || !token} className="w-full">
          {loading ? "Resetting..." : "Reset password"}
        </Button>
        {!token ? (
          <p className="text-center text-sm text-destructive">
            Missing reset token. Please open the link from your email.
          </p>
        ) : null}
        <p className="text-center text-sm text-muted-foreground">
          <Link
            href="/login"
            className="text-foreground font-medium underline underline-offset-4"
          >
            Back to login
          </Link>
        </p>
      </div>
    </form>
  );
}
