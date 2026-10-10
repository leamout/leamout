"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const schema = z.object({
  host: z
    .string()
    .trim()
    .min(1, "Enter a hostname or IP address.")
    .refine(
      (value) => !/[\s/]/.test(value),
      "Enter a host without a scheme or path.",
    ),
  port: z
    .string()
    .refine(
      (value) =>
        /^\d+$/.test(value) && Number(value) >= 1 && Number(value) <= 65535,
      "Enter a port between 1 and 65535.",
    ),
  transport: z.enum(["udp", "tcp", "tls"]),
});

export function TrunkConnectionForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { host: "", port: "5060", transport: "udp" },
  });

  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice("Preview only. No endpoint was added."),
      )}
    >
      <h2 className="text-lg font-semibold">Add endpoint</h2>
      <p className="text-sm text-muted-foreground">
        Endpoint health and existing endpoints have not been loaded.
      </p>
      {(["host", "port", "transport"] as const).map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`endpoint-${name}`}>
                {name === "host"
                  ? "Hostname or IP address"
                  : name === "port"
                    ? "Port"
                    : "Transport"}
              </FieldLabel>
              {name === "transport" ? (
                <select
                  {...field}
                  id={`endpoint-${name}`}
                  className="h-9 rounded-md border bg-background px-3"
                >
                  {["udp", "tcp", "tls"].map((value) => (
                    <option key={value} value={value}>
                      {value.toUpperCase()}
                    </option>
                  ))}
                </select>
              ) : (
                <Input
                  {...field}
                  id={`endpoint-${name}`}
                  type={name === "port" ? "number" : "text"}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `endpoint-${name}-error` : undefined
                  }
                />
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`endpoint-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
      ))}
      <Button type="submit">Add endpoint</Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
