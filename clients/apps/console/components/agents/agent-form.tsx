"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { Textarea } from "@leamout/ui/components/textarea";
import Link from "next/link";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

const schema = z.object({
  name: z.string().trim().min(1, "Enter an agent name."),
  description: z.string(),
  instructions: z.string(),
  greeting: z.string(),
  language: z.string(),
});

export function AgentForm({ creating = false }: { creating?: boolean }) {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      description: "",
      instructions: "",
      greeting: "",
      language: "",
    },
  });
  const fields = creating
    ? (["name", "description"] as const)
    : ([
        "name",
        "description",
        "instructions",
        "greeting",
        "language",
      ] as const);
  const labels = {
    name: "Name",
    description: "Description (optional)",
    instructions: "Instructions",
    greeting: "Greeting",
    language: "Language",
  };

  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice(
          creating
            ? "Preview only. No agent was created."
            : "Preview only. Changes were not saved.",
        ),
      )}
    >
      {fields.map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`agent-${name}`}>{labels[name]}</FieldLabel>
              {name === "description" || name === "instructions" ? (
                <Textarea
                  {...field}
                  id={`agent-${name}`}
                  rows={name === "instructions" ? 6 : 3}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `agent-${name}-error` : undefined
                  }
                />
              ) : (
                <Input
                  {...field}
                  id={`agent-${name}`}
                  required={name === "name"}
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `agent-${name}-error` : undefined
                  }
                />
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`agent-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
      ))}
      <div className="flex gap-3">
        <Button type="submit">
          {creating ? "Create agent" : "Save changes"}
        </Button>
        {creating && (
          <Button variant="outline" render={<Link href="/agents" />}>
            Cancel
          </Button>
        )}
      </div>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
