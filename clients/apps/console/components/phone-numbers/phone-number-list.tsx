"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@leamout/ui/components/table";
import { useQuery } from "@tanstack/react-query";
import Link from "next/link";
import { useState } from "react";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";
import type { PhoneNumber } from "./types";

export function PhoneNumberList() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const [search, setSearch] = useState("");
  const [voice, setVoice] = useState("all");
  const query = useQuery({
    queryKey: ["numbers", organizationId],
    queryFn: ({ signal }) => api<{ numbers: PhoneNumber[] }>("/numbers/", { organizationId, signal }),
    enabled: Boolean(organizationId),
  });
  const numbers = query.data?.numbers ?? [];
  const filtered = numbers.filter((number) =>
    (number.number + " " + number.country_code).toLowerCase().includes(search.trim().toLowerCase()) &&
    (voice === "all" || number.voice_enabled === (voice === "enabled")),
  );
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div>
          <h1 className="text-3xl font-semibold tracking-tight">Phone numbers</h1>
          <p className="mt-2 text-muted-foreground">Register existing carrier numbers and configure inbound routing.</p>
        </div>
        <Button render={<Link href="/phone-numbers/new" />}>Connect phone number</Button>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="flex-1 space-y-2">
          <label className="text-sm font-medium" htmlFor="number-search">Search numbers</label>
          <Input id="number-search" type="search" value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
        <label className="space-y-2 text-sm">Voice
          <select className="block h-9 rounded-md border bg-background px-3" value={voice} onChange={(e) => setVoice(e.target.value)}>
            <option value="all">All</option><option value="enabled">Enabled</option><option value="disabled">Disabled</option>
          </select>
        </label>
      </div>
      {query.isPending ? <p role="status">Loading phone numbers…</p> : query.isError ? (
        <div role="alert" className="space-y-2"><p>{errorMessage(query.error)}</p><Button variant="outline" onClick={() => void query.refetch()}>Retry</Button></div>
      ) : (
        <div className="overflow-hidden rounded-lg border">
          <Table><TableHeader><TableRow>
            {["Number", "Country", "Status", "Voice", "SIP trunk"].map((title) => <TableHead key={title}>{title}</TableHead>)}
          </TableRow></TableHeader><TableBody>
            {filtered.map((number) => (
              <TableRow key={number.id}>
                <TableCell><Link href={"/phone-numbers/" + encodeURIComponent(number.id)} className="font-medium hover:underline">{number.number}</Link></TableCell>
                <TableCell>{number.country_code}</TableCell><TableCell>{number.status}</TableCell>
                <TableCell>{number.voice_enabled ? "Enabled" : "Disabled"}</TableCell>
                <TableCell>{number.trunk_id ? <Link className="hover:underline" href={"/sip-trunks/" + encodeURIComponent(number.trunk_id)}>{number.trunk_id}</Link> : "Unassigned"}</TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && <TableRow><TableCell colSpan={5} className="py-12 text-center text-muted-foreground">
              {numbers.length ? "No matching phone numbers." : "No numbers registered yet."}
            </TableCell></TableRow>}
          </TableBody></Table>
        </div>
      )}
      <p className="text-xs text-muted-foreground">Phone number purchasing and provisioning are not included. Register only numbers you control.</p>
    </div>
  );
}
