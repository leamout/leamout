"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const positiveInteger = z
  .string()
  .refine(
    (value) =>
      value === "" ||
      (/^\d+$/.test(value) &&
        Number(value) >= 1 &&
        Number(value) <= 2147483647),
    "Enter a positive integer.",
  );
const schema = z.object({
  name: z.string().trim().min(1, "Enter a trunk name.").max(255),
  direction: z.enum(["inbound", "outbound", "bidirectional"]),
  status: z.enum(["active", "disabled"]),
  maxCps: positiveInteger,
  maxConcurrentCalls: positiveInteger,
});

export function TrunkForm({ creating = false }: { creating?: boolean }) {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      direction: "bidirectional",
      status: "disabled",
      maxCps: "",
      maxConcurrentCalls: "",
    },
  });
  const labels = {
    name: "Name",
    direction: "Direction",
    status: "Status",
    maxCps: "Maximum calls per second",
    maxConcurrentCalls: "Maximum concurrent calls",
  };
  const choices = {
    direction: ["inbound", "outbound", "bidirectional"],
    status: ["active", "disabled"],
  };

  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice(
          creating
            ? "Preview only. No SIP trunk was created."
            : "Preview only. Changes were not saved.",
        ),
      )}
    >
      {(Object.keys(labels) as (keyof typeof labels)[]).map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`trunk-${name}`}>{labels[name]}</FieldLabel>
              {name === "direction" || name === "status" ? (
                <select
                  {...field}
                  id={`trunk-${name}`}
                  className="h-9 rounded-md border bg-background px-3 text-sm"
                >
                  {choices[name].map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              ) : (
                <Input
                  {...field}
                  id={`trunk-${name}`}
                  type={name === "name" ? "text" : "number"}
                  min={name === "name" ? undefined : 1}
                  step={name === "name" ? undefined : 1}
                  required={name === "name"}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `trunk-${name}-error` : undefined
                  }
                />
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`trunk-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
      ))}
      <p className="text-sm text-muted-foreground">
        Endpoints and authentication are configured in trunk details. Blank
        capacity limits use deployment defaults when integrated.
      </p>
      <Button type="submit">
        {creating ? "Create SIP trunk" : "Save changes"}
      </Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
