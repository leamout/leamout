"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldLabel,
} from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const schema = z.object({
  name: z.string().trim().min(1, "Enter a credential name."),
  secret: z.string().trim().min(1, "Enter an API key."),
});

export function ProviderCredentialForm() {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", secret: "" },
  });
  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() => {
        form.resetField("secret");
        setNotice(
          "Preview only. Credentials were not saved or verified. The API key field has been cleared.",
        );
      })}
    >
      <h2 className="text-lg font-semibold">Add organization credential</h2>
      {(["name", "secret"] as const).map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`provider-${name}`}>
                {name === "name" ? "Credential name" : "API key"}
              </FieldLabel>
              <Input
                {...field}
                id={`provider-${name}`}
                type={name === "secret" ? "password" : "text"}
                autoComplete={name === "secret" ? "new-password" : "off"}
                required
                aria-invalid={fieldState.invalid}
                aria-describedby={
                  fieldState.invalid ? `provider-${name}-error` : undefined
                }
              />
              {fieldState.invalid && (
                <FieldError
                  id={`provider-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
              {name === "secret" && (
                <FieldDescription>
                  Preview input only. The key is not sent to a server or stored
                  in browser storage.
                </FieldDescription>
              )}
            </Field>
          )}
        />
      ))}
      <Button type="submit">Save credential</Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
