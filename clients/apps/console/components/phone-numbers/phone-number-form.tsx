"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { useRef } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { SipTrunk } from "@/components/sip-trunks/types";
import type { PhoneNumber } from "./types";

const schema = z.object({
  number: z.string().trim().regex(/^\+[1-9][0-9]{6,14}$/, "Use E.164 format, e.g. +233201234567."),
  countryCode: z.string().trim().regex(/^[A-Za-z]{2}$/, "Use a two-letter country code."),
  trunkId: z.string(),
  voiceEnabled: z.boolean(),
});

type Values = z.infer<typeof schema>;

export function PhoneNumberForm({ creating = false, number }: { creating?: boolean; number?: PhoneNumber }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const idempotencyKey = useRef<string | null>(null);
  const trunks = useQuery({
    queryKey: ["trunks", organizationId],
    queryFn: ({ signal }) => api<{ trunks: SipTrunk[] }>("/trunks/", { organizationId, signal }),
    enabled: Boolean(organizationId),
  });
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      number: number?.number ?? "",
      countryCode: number?.country_code ?? "",
      trunkId: number?.trunk_id ?? "",
      voiceEnabled: number?.voice_enabled ?? false,
    },
  });
  const save = useMutation({
    mutationFn: async (values: Values) => {
      const path = "/numbers/" + encodeURIComponent(number?.id ?? "");
      if (creating) {
        idempotencyKey.current ??= crypto.randomUUID();
        return api<PhoneNumber>("/numbers/", {
          method: "POST", organizationId, idempotencyKey: idempotencyKey.current,
          body: {
            number: values.number.trim(),
            country_code: values.countryCode.trim().toUpperCase(),
            ...(values.trunkId ? { trunk_id: values.trunkId } : {}),
            voice_enabled: values.voiceEnabled,
          },
        });
      }
      if (!number) throw new Error("Phone number not loaded.");
      // Reject unsupported unassignment before issuing any writes.
      if (!values.trunkId && number.trunk_id) {
        throw new Error("The backend does not support clearing the assigned trunk.");
      }
      // The backend uses separate endpoints for voice capability and trunk assignment.
      let result = number;
      if (values.voiceEnabled !== number.voice_enabled) {
        result = await api<PhoneNumber>(path, {
          method: "PATCH", organizationId, body: { voice_enabled: values.voiceEnabled },
        });
      }
      if (values.trunkId && values.trunkId !== number.trunk_id) {
        result = await api<PhoneNumber>(path + "/trunk", {
          method: "PATCH", organizationId, body: { trunk_id: values.trunkId },
        });
      }
      return result;
    },
    onSuccess: async (saved) => {
      client.setQueryData(["numbers", organizationId, saved.id], saved);
      await client.invalidateQueries({ queryKey: ["numbers", organizationId] });
      if (creating) router.push("/phone-numbers/" + encodeURIComponent(saved.id));
      else form.reset({
        number: saved.number, countryCode: saved.country_code,
        trunkId: saved.trunk_id ?? "", voiceEnabled: saved.voice_enabled,
      });
    },
  });
  return (
    <form className="max-w-2xl space-y-5" noValidate
      onChange={() => { if (creating) idempotencyKey.current = null; }}
      onSubmit={form.handleSubmit((values) => save.mutate(values))}>
      <div className="grid gap-4 sm:grid-cols-2">
        <Controller control={form.control} name="number" render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}><FieldLabel htmlFor="number-e164">Phone number (E.164)</FieldLabel>
            <Input {...field} id="number-e164" type="tel" readOnly={!creating} required placeholder="+233201234567" />
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </Field>
        )} />
        <Controller control={form.control} name="countryCode" render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}><FieldLabel htmlFor="number-country">Country code</FieldLabel>
            <Input {...field} id="number-country" readOnly={!creating} required maxLength={2} placeholder="GH" />
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </Field>
        )} />
      </div>
      <Field><FieldLabel htmlFor="number-trunk">SIP trunk</FieldLabel>
        <select id="number-trunk" className="h-10 w-full rounded-md border bg-background px-3" {...form.register("trunkId")}>
          <option value="">Unassigned</option>
          {trunks.data?.trunks.map((trunk) => <option key={trunk.id} value={trunk.id}>{trunk.name}</option>)}
        </select>
        {trunks.isError && <p role="alert" className="text-sm text-destructive">{errorMessage(trunks.error)}</p>}
        {number?.trunk_id && <p className="text-xs text-muted-foreground">A trunk can be changed, but not cleared using the current API.</p>}
      </Field>
      <label className="flex items-center gap-3 text-sm"><input type="checkbox" {...form.register("voiceEnabled")} /> Enable voice</label>
      <Button type="submit" disabled={save.isPending || !organizationId || trunks.isPending || trunks.isError}>
        {save.isPending ? "Saving…" : creating ? "Connect existing number" : "Save changes"}
      </Button>
      {save.isSuccess && !creating && <p role="status" className="text-sm">Phone number updated.</p>}
      {save.isError && <p role="alert" className="text-sm text-destructive">{errorMessage(save.error)}</p>}
    </form>
  );
}
