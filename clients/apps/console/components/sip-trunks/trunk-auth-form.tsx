"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const schema = z
  .object({
    method: z.enum(["ip", "none", "digest"]),
    username: z.string().trim(),
    realm: z.string().trim(),
    secret: z.string(),
    cidr: z.string(),
  })
  .superRefine((data, context) => {
    if (
      data.method === "ip" &&
      !z.union([z.cidrv4(), z.cidrv6()]).safeParse(data.cidr.trim()).success
    ) {
      context.addIssue({
        code: "custom",
        path: ["cidr"],
        message: "Enter a valid IPv4 or IPv6 network prefix.",
      });
    }
    if (data.method === "digest") {
      for (const name of ["username", "realm", "secret"] as const) {
        if (!data[name])
          context.addIssue({
            code: "custom",
            path: [name],
            message: `Enter a ${name}.`,
          });
      }
    }
  });

export function TrunkAuthForm({ inbound }: { inbound: boolean }) {
  const [notice, setNotice] = useState("");
  const prefix = inbound ? "inbound" : "outbound";
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      method: inbound ? "ip" : "none",
      username: "",
      realm: "",
      secret: "",
      cidr: "",
    },
  });
  const method = form.watch("method");

  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() => {
        setNotice("Preview only. Authentication was not saved.");
        form.setValue("secret", "");
      })}
    >
      <h2 className="text-lg font-semibold">
        {inbound ? "Inbound" : "Outbound"} authentication
      </h2>
      <Controller
        name="method"
        control={form.control}
        render={({ field }) => (
          <Field>
            <FieldLabel htmlFor={`${prefix}-method`}>Method</FieldLabel>
            <select
              {...field}
              id={`${prefix}-method`}
              className="h-9 rounded-md border bg-background px-3"
            >
              <option value={inbound ? "ip" : "none"}>
                {inbound ? "Source IP allowlist" : "No digest authentication"}
              </option>
              <option value="digest">Digest authentication</option>
            </select>
          </Field>
        )}
      />
      {method === "digest" &&
        (["username", "realm", "secret"] as const).map((name) => (
          <Controller
            key={name}
            name={name}
            control={form.control}
            render={({ field, fieldState }) => (
              <Field data-invalid={fieldState.invalid}>
                <FieldLabel htmlFor={`${prefix}-${name}`}>
                  {name === "secret"
                    ? "Password"
                    : name === "realm"
                      ? "Realm"
                      : "Username"}
                </FieldLabel>
                <Input
                  {...field}
                  id={`${prefix}-${name}`}
                  type={name === "secret" ? "password" : "text"}
                  autoComplete={name === "secret" ? "new-password" : "off"}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `${prefix}-${name}-error` : undefined
                  }
                />
                {fieldState.invalid && (
                  <FieldError
                    id={`${prefix}-${name}-error`}
                    errors={[fieldState.error]}
                  />
                )}
              </Field>
            )}
          />
        ))}
      {method === "ip" && (
        <Controller
          name="cidr"
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor="inbound-cidr">
                Allowed source network (CIDR)
              </FieldLabel>
              <Input
                {...field}
                id="inbound-cidr"
                placeholder="203.0.113.0/24"
                aria-invalid={fieldState.invalid}
                aria-describedby={
                  fieldState.invalid ? "inbound-cidr-error" : undefined
                }
              />
              {fieldState.invalid && (
                <FieldError
                  id="inbound-cidr-error"
                  errors={[fieldState.error]}
                />
              )}
              <p className="text-sm text-muted-foreground">
                Preview a source network. Existing allowlist entries have not
                been loaded.
              </p>
            </Field>
          )}
        />
      )}
      <Button type="submit">Save authentication</Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
