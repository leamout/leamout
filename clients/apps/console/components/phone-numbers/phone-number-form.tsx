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
import Link from "next/link";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import type {
  PhoneNumber,
  ResourceOption,
} from "@/components/phone-numbers/types";

const schema = z.object({
  number: z
    .string()
    .trim()
    .regex(
      /^\+[1-9][0-9]{6,14}$/,
      "Enter a phone number in E.164 format, such as +233201234567.",
    ),
  countryCode: z
    .string()
    .trim()
    .regex(/^[A-Za-z]{2}$/, "Enter a two-letter country code, such as GH."),
  trunkId: z.string(),
  voiceEnabled: z.boolean(),
});

export function PhoneNumberForm({
  number,
  creating = false,
  trunks = [],
}: {
  number?: PhoneNumber;
  creating?: boolean;
  trunks?: ResourceOption[];
}) {
  const [notice, setNotice] = useState("");
  const options =
    number?.trunk && !trunks.some((trunk) => trunk.id === number.trunk?.id)
      ? [number.trunk, ...trunks]
      : trunks;
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      number: number?.number ?? "",
      countryCode: number?.countryCode ?? "",
      trunkId: number?.trunk?.id ?? "",
      voiceEnabled: number?.voiceEnabled ?? false,
    },
  });
  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice(
          creating
            ? "Preview only. No phone number was connected."
            : "Preview only. Changes were not saved.",
        ),
      )}
    >
      {(["number", "countryCode"] as const).map((name) => (
        <Controller
          key={name}
          name={name}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`phone-${name}`}>
                {name === "number" ? "Phone number" : "Country code"}
              </FieldLabel>
              <Input
                {...field}
                id={`phone-${name}`}
                type={name === "number" ? "tel" : "text"}
                readOnly={!creating}
                placeholder={name === "number" ? "+233201234567" : "GH"}
                aria-invalid={fieldState.invalid}
                aria-describedby={
                  fieldState.invalid ? `phone-${name}-error` : undefined
                }
              />
              {fieldState.invalid && (
                <FieldError
                  id={`phone-${name}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
      ))}
      <Controller
        name="trunkId"
        control={form.control}
        render={({ field }) => (
          <Field>
            <FieldLabel htmlFor="phone-trunk">SIP trunk</FieldLabel>
            <select
              {...field}
              id="phone-trunk"
              className="h-9 rounded-md border bg-background px-3"
            >
              <option value="">Unassigned</option>
              {options.map((trunk) => (
                <option key={trunk.id} value={trunk.id}>
                  {trunk.name}
                </option>
              ))}
            </select>
            <FieldDescription>
              {options.length
                ? "Choose the trunk carrying this existing number."
                : "Trunk options have not been loaded."}{" "}
              <Link href="/sip-trunks" className="underline">
                Manage SIP trunks
              </Link>
            </FieldDescription>
          </Field>
        )}
      />
      <Controller
        name="voiceEnabled"
        control={form.control}
        render={({ field }) => (
          <Field>
            <label
              htmlFor="phone-voice"
              className="flex items-center gap-3 text-sm font-medium"
            >
              <input
                id="phone-voice"
                type="checkbox"
                name={field.name}
                ref={field.ref}
                onBlur={field.onBlur}
                checked={field.value}
                onChange={(event) => field.onChange(event.target.checked)}
              />
              Enable voice
            </label>
          </Field>
        )}
      />
      <Button type="submit" disabled={!creating && !number}>
        {creating ? "Connect phone number" : "Save changes"}
      </Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
