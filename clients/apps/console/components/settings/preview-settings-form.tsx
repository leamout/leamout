"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { Textarea } from "@leamout/ui/components/textarea";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

type FormField = {
  name: string;
  label: string;
  type?: "text" | "email" | "password" | "url" | "textarea";
  required?: boolean;
  options?: string[];
};

export function PreviewSettingsForm({
  fields,
  action = "Save changes",
  validate,
}: {
  fields: FormField[];
  action?: string;
  validate?: (
    values: Record<string, string>,
  ) => { field: string; message: string }[];
}) {
  const [notice, setNotice] = useState("");
  const schema = z
    .object(
      Object.fromEntries(
        fields.map((field) => [
          field.name,
          field.required
            ? z
                .string()
                .min(1, `Enter ${field.label.toLowerCase()}.`)
                .refine(
                  (value) =>
                    field.type === "password" || value.trim().length > 0,
                  "This field is required.",
                )
            : z.string(),
        ]),
      ),
    )
    .superRefine((values, context) => {
      for (const field of fields) {
        const value = values[field.name];
        if (
          value &&
          field.type === "email" &&
          !z.email().safeParse(value.trim()).success
        )
          context.addIssue({
            code: "custom",
            path: [field.name],
            message: "Enter a valid email address.",
          });
        if (
          value &&
          field.type === "url" &&
          !z.url().safeParse(value.trim()).success
        )
          context.addIssue({
            code: "custom",
            path: [field.name],
            message: "Enter a valid URL.",
          });
      }
      for (const error of validate?.(values) ?? [])
        context.addIssue({
          code: "custom",
          path: [error.field],
          message: error.message,
        });
    });
  const form = useForm<Record<string, string>>({
    resolver: zodResolver(schema),
    defaultValues: Object.fromEntries(fields.map((field) => [field.name, ""])),
  });
  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() => {
        for (const field of fields)
          if (field.type === "password") form.resetField(field.name);
        setNotice(
          "Preview only. Nothing was saved, sent, or generated. Sensitive fields have been cleared.",
        );
      })}
    >
      {fields.map((item) => (
        <Controller
          key={item.name}
          name={item.name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`settings-${item.name}`}>
                {item.label}
              </FieldLabel>
              {item.options ? (
                <select
                  {...field}
                  id={`settings-${item.name}`}
                  className="h-9 rounded-md border bg-background px-3"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? `settings-${item.name}-error`
                      : undefined
                  }
                >
                  <option value="">Choose a role</option>
                  {item.options.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              ) : item.type === "textarea" ? (
                <Textarea
                  {...field}
                  id={`settings-${item.name}`}
                  rows={4}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? `settings-${item.name}-error`
                      : undefined
                  }
                />
              ) : (
                <Input
                  {...field}
                  id={`settings-${item.name}`}
                  type={item.type ?? "text"}
                  autoComplete={
                    item.type === "password" ? "new-password" : "off"
                  }
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid
                      ? `settings-${item.name}-error`
                      : undefined
                  }
                />
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`settings-${item.name}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
      ))}
      <Button type="submit">{action}</Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
