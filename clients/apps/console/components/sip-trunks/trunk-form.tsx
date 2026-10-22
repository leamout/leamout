"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { SipTrunk } from "./types";

const limit = z.string().refine(
  (value) => value === "" || (/^\d+$/.test(value) && Number(value) > 0 && Number(value) <= 2147483647),
  "Enter a positive integer.",
);

const schema = z.object({
  name: z.string().trim().min(1).max(255),
  direction: z.enum(["inbound", "outbound", "bidirectional"]),
  status: z.enum(["active", "disabled"]),
  inboundEnabled: z.boolean(),
  maxCps: limit,
  maxConcurrentCalls: limit,
});

type Values = z.infer<typeof schema>;

export function TrunkForm({ creating = false, trunk }: { creating?: boolean; trunk?: SipTrunk }) {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const client = useQueryClient();
  const router = useRouter();
  const form = useForm<Values>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: trunk?.name ?? "",
      direction: trunk?.direction ?? "bidirectional",
      status: trunk?.status ?? "disabled",
      inboundEnabled: trunk?.inbound_enabled ?? false,
      maxCps: trunk ? String(trunk.max_cps) : "",
      maxConcurrentCalls: trunk ? String(trunk.max_concurrent_calls) : "",
    },
  });
  const save = useMutation({
    mutationFn: (values: Values) => {
      const body: Record<string, unknown> = {
        name: values.name.trim(),
        direction: values.direction,
        status: values.status,
        inbound_enabled: values.inboundEnabled,
      };
      if (values.maxCps) body.max_cps = Number(values.maxCps);
      if (values.maxConcurrentCalls) body.max_concurrent_calls = Number(values.maxConcurrentCalls);
      return api<SipTrunk>(
        creating ? "/trunks/" : "/trunks/" + encodeURIComponent(trunk?.id ?? ""),
        { method: creating ? "POST" : "PATCH", organizationId, body },
      );
    },
    onSuccess: async (saved) => {
      client.setQueryData(["trunks", organizationId, saved.id], saved);
      await client.invalidateQueries({ queryKey: ["trunks", organizationId] });
      if (creating) router.push("/sip-trunks/" + encodeURIComponent(saved.id));
      else form.reset({
        name: saved.name,
        direction: saved.direction,
        status: saved.status,
        inboundEnabled: saved.inbound_enabled,
        maxCps: String(saved.max_cps),
        maxConcurrentCalls: String(saved.max_concurrent_calls),
      });
    },
  });
  return (
    <form className="max-w-2xl space-y-5" noValidate onSubmit={form.handleSubmit((values) => save.mutate(values))}>
      <Controller name="name" control={form.control} render={({ field, fieldState }) => (
        <Field data-invalid={fieldState.invalid}>
          <FieldLabel htmlFor="trunk-name">Trunk name</FieldLabel>
          <Input {...field} id="trunk-name" required maxLength={255} />
          {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
        </Field>
      )} />
      <div className="grid gap-4 sm:grid-cols-2">
        <Field><FieldLabel htmlFor="trunk-direction">Direction</FieldLabel>
          <select id="trunk-direction" {...form.register("direction")} className="h-10 w-full rounded-md border bg-background px-3">
            <option value="inbound">Inbound</option>
            <option value="outbound">Outbound</option>
            <option value="bidirectional">Bidirectional</option>
          </select>
        </Field>
        <Field><FieldLabel htmlFor="trunk-status">Status</FieldLabel>
          <select id="trunk-status" {...form.register("status")} className="h-10 w-full rounded-md border bg-background px-3">
            <option value="disabled">Disabled</option>
            <option value="active">Active</option>
          </select>
        </Field>
      </div>
      <label className="flex items-center gap-3 text-sm">
        <input type="checkbox" {...form.register("inboundEnabled")} />
        Allow inbound calls
      </label>
      <div className="grid gap-4 sm:grid-cols-2">
        <Controller name="maxCps" control={form.control} render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="trunk-cps">Max calls per second</FieldLabel>
            <Input {...field} id="trunk-cps" type="number" min={1} step={1} />
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </Field>
        )} />
        <Controller name="maxConcurrentCalls" control={form.control} render={({ field, fieldState }) => (
          <Field data-invalid={fieldState.invalid}>
            <FieldLabel htmlFor="trunk-concurrent">Max concurrent calls</FieldLabel>
            <Input {...field} id="trunk-concurrent" type="number" min={1} step={1} />
            {fieldState.invalid && <FieldError errors={[fieldState.error]} />}
          </Field>
        )} />
      </div>
      <p className="text-xs text-muted-foreground">Configure SIP endpoints and credentials after saving. Blank limits use server defaults on creation.</p>
      <Button type="submit" disabled={save.isPending || !organizationId}>
        {save.isPending ? "Saving…" : creating ? "Create SIP trunk" : "Save changes"}
      </Button>
      {save.isSuccess && !creating && <p role="status" className="text-sm">Trunk updated.</p>}
      {save.isError && <p role="alert" className="text-sm text-destructive">{errorMessage(save.error)}</p>}
    </form>
  );
}
