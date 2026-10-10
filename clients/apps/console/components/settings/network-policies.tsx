"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";

export type NetworkPolicy = {
  id: string;
  name: string;
  action: "allow" | "deny";
  sourceCIDR: string;
  status: "active" | "disabled";
};

const schema = z.object({
  name: z
    .string()
    .trim()
    .min(1, "Enter a policy name.")
    .max(128, "Use at most 128 characters."),
  action: z.enum(["allow", "deny"]),
  sourceCIDR: z
    .string()
    .trim()
    .refine(
      (value) =>
        z.cidrv4().safeParse(value).success ||
        z.cidrv6().safeParse(value).success,
      "Enter a valid IPv4 or IPv6 CIDR, such as 203.0.113.0/24 or 2001:db8::/32.",
    ),
  status: z.enum(["active", "disabled"]),
});

function PolicyForm({ policy }: { policy?: NetworkPolicy }) {
  const [notice, setNotice] = useState("");
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: policy?.name ?? "",
      action: policy?.action ?? "allow",
      sourceCIDR: policy?.sourceCIDR ?? "",
      status: policy?.status ?? "active",
    },
  });
  const fields = [
    { name: "name", label: "Policy name" },
    { name: "sourceCIDR", label: "Source CIDR" },
    { name: "action", label: "Action", options: ["allow", "deny"] },
    ...(policy
      ? [{ name: "status", label: "Status", options: ["active", "disabled"] }]
      : []),
  ] as const;

  return (
    <form
      className="max-w-2xl space-y-6"
      noValidate
      onChange={() => setNotice("")}
      onSubmit={form.handleSubmit(() =>
        setNotice("Preview only. No network policy was saved or enforced."),
      )}
    >
      {fields.map((item) => (
        <Controller
          key={item.name}
          name={item.name as keyof z.infer<typeof schema>}
          control={form.control}
          render={({ field, fieldState }) => (
            <Field data-invalid={fieldState.invalid}>
              <FieldLabel htmlFor={`policy-${item.name}`}>
                {item.label}
              </FieldLabel>
              {"options" in item ? (
                <select
                  {...field}
                  id={`policy-${item.name}`}
                  className="h-9 rounded-md border bg-background px-3"
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `policy-${item.name}-error` : undefined
                  }
                >
                  {item.options.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              ) : (
                <Input
                  {...field}
                  id={`policy-${item.name}`}
                  placeholder={
                    item.name === "sourceCIDR"
                      ? "203.0.113.0/24"
                      : "Office network"
                  }
                  aria-invalid={fieldState.invalid}
                  aria-describedby={
                    fieldState.invalid ? `policy-${item.name}-error` : undefined
                  }
                />
              )}
              {fieldState.invalid && (
                <FieldError
                  id={`policy-${item.name}-error`}
                  errors={[fieldState.error]}
                />
              )}
            </Field>
          )}
        />
      ))}
      <Button type="submit">
        {policy ? "Preview changes" : "Preview policy creation"}
      </Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}

export function NetworkPolicies({
  policies = [],
  available,
}: {
  policies?: NetworkPolicy[];
  available?: boolean;
}) {
  const [selected, setSelected] = useState<NetworkPolicy>();
  const [search, setSearch] = useState("");
  const [deleting, setDeleting] = useState<NetworkPolicy>();
  const [notice, setNotice] = useState("");
  const filtered = policies.filter((policy) =>
    `${policy.name} ${policy.sourceCIDR}`
      .toLowerCase()
      .includes(search.trim().toLowerCase()),
  );

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Network policies
        </h1>
        <p className="text-muted-foreground">
          Manage source IP rules for organization API access.
        </p>
      </div>
      {available === false ? (
        <p className="rounded-lg border p-6 text-sm text-muted-foreground">
          Network policies require the private networking capability for your
          organization.
        </p>
      ) : (
        <>
          <div className="space-y-2">
            <label htmlFor="policy-search" className="text-sm font-medium">
              Search policies
            </label>
            <Input
              id="policy-search"
              type="search"
              placeholder="Policy name or CIDR"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
            />
          </div>
          <div className="overflow-hidden rounded-lg border">
            <Table>
              <TableHeader>
                <TableRow>
                  {["Name", "Source CIDR", "Action", "Status", "Actions"].map(
                    (label) => (
                      <TableHead key={label}>{label}</TableHead>
                    ),
                  )}
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.map((policy) => (
                  <TableRow key={policy.id}>
                    <TableCell className="font-medium">{policy.name}</TableCell>
                    <TableCell className="font-mono text-xs">
                      {policy.sourceCIDR}
                    </TableCell>
                    <TableCell className="capitalize">
                      {policy.action}
                    </TableCell>
                    <TableCell className="capitalize">
                      {policy.status}
                    </TableCell>
                    <TableCell>
                      <div className="flex gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`Edit ${policy.name}`}
                          onClick={() => {
                            setSelected(policy);
                            setDeleting(undefined);
                            setNotice("");
                          }}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          aria-label={`Delete ${policy.name}`}
                          onClick={() => {
                            setDeleting(policy);
                            setNotice("");
                          }}
                        >
                          Delete
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {filtered.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={5} className="h-32 text-center">
                      <p className="font-medium">
                        {policies.length === 0
                          ? "No network policies to display"
                          : "No matching policies"}
                      </p>
                      <p className="mt-2 text-muted-foreground">
                        {policies.length === 0
                          ? "Configured network rules will appear here."
                          : "Try another search."}
                      </p>
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
          {deleting && (
            <div className="space-y-3 rounded-lg border p-6">
              <p className="font-medium">
                Preview deletion of {deleting.name}?
              </p>
              <div className="flex gap-3">
                <Button
                  variant="destructive"
                  onClick={() => {
                    setDeleting(undefined);
                    setNotice("Deletion preview only. No policy was removed.");
                  }}
                >
                  Confirm deletion preview
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setDeleting(undefined)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          )}
          <p role="status" className="text-sm text-muted-foreground">
            {notice}
          </p>
          <section
            className="space-y-6 rounded-lg border p-6"
            aria-labelledby="policy-form-title"
          >
            <div className="flex flex-wrap items-center justify-between gap-3">
              <h2 id="policy-form-title" className="font-semibold">
                {selected ? `Edit ${selected.name}` : "Create network policy"}
              </h2>
              {selected && (
                <Button
                  variant="outline"
                  onClick={() => setSelected(undefined)}
                >
                  Cancel editing
                </Button>
              )}
            </div>
            <PolicyForm key={selected?.id ?? "new"} policy={selected} />
          </section>
        </>
      )}
      <p className="text-sm text-muted-foreground">
        UI preview. Policies and organization capabilities are not connected.
      </p>
    </div>
  );
}
