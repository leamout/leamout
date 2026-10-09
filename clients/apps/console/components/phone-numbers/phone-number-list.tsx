"use client";

import { Button } from "@leamout/ui/components/button";
import { Input } from "@leamout/ui/components/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@leamout/ui/components/table";
import Link from "next/link";
import { useState } from "react";
import type { PhoneNumber } from "@/components/phone-numbers/types";

export function PhoneNumberList({ numbers = [] }: { numbers?: PhoneNumber[] }) {
  const [search, setSearch] = useState("");
  const [voice, setVoice] = useState("all");
  const filtered = numbers.filter(
    (number) =>
      `${number.number} ${number.countryCode} ${number.trunk?.name ?? ""} ${number.agent?.name ?? ""}`
        .toLowerCase()
        .includes(search.toLowerCase()) &&
      (voice === "all" || number.voiceEnabled === (voice === "enabled")),
  );
  return (
    <div className="space-y-6">
      <div className="flex flex-wrap justify-between gap-4">
        <div className="space-y-2">
          <h1 className="text-3xl font-semibold tracking-tight">
            Phone numbers
          </h1>
          <p className="text-muted-foreground">
            Manage your existing BYOC numbers and agent routing.
          </p>
        </div>
        <Button render={<Link href="/phone-numbers/new" />}>
          Connect phone number
        </Button>
      </div>
      <div className="flex flex-wrap gap-4">
        <div className="flex-1 space-y-2">
          <label htmlFor="number-search" className="text-sm font-medium">
            Search numbers
          </label>
          <Input
            id="number-search"
            type="search"
            placeholder="Number, country, trunk, or agent"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        <div className="space-y-2">
          <label
            htmlFor="number-voice-filter"
            className="block text-sm font-medium"
          >
            Voice
          </label>
          <select
            id="number-voice-filter"
            className="h-9 rounded-md border bg-background px-3"
            value={voice}
            onChange={(event) => setVoice(event.target.value)}
          >
            <option value="all">All numbers</option>
            <option value="enabled">Enabled</option>
            <option value="disabled">Disabled</option>
          </select>
        </div>
      </div>
      <div className="overflow-hidden rounded-lg border">
        <Table>
          <TableHeader>
            <TableRow>
              {[
                "Number",
                "Country",
                "Status",
                "Voice",
                "SIP trunk",
                "Agent",
              ].map((label) => (
                <TableHead key={label}>{label}</TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {filtered.map((number) => (
              <TableRow key={number.id}>
                <TableCell>
                  <Link
                    className="font-medium hover:underline"
                    href={`/phone-numbers/${encodeURIComponent(number.id)}`}
                  >
                    {number.number}
                  </Link>
                </TableCell>
                <TableCell>{number.countryCode}</TableCell>
                <TableCell>{number.status}</TableCell>
                <TableCell>
                  {number.voiceEnabled ? "Enabled" : "Disabled"}
                </TableCell>
                <TableCell>
                  {number.trunk ? (
                    <Link
                      className="hover:underline"
                      href={`/sip-trunks/${encodeURIComponent(number.trunk.id)}`}
                    >
                      {number.trunk.name}
                    </Link>
                  ) : (
                    "Unassigned"
                  )}
                </TableCell>
                <TableCell>
                  {number.agent ? (
                    <Link
                      className="hover:underline"
                      href={`/agents/${encodeURIComponent(number.agent.id)}`}
                    >
                      {number.agent.name}
                    </Link>
                  ) : (
                    "Unassigned"
                  )}
                </TableCell>
              </TableRow>
            ))}
            {filtered.length === 0 && (
              <TableRow>
                <TableCell colSpan={6} className="h-40 text-center">
                  <p className="font-medium">
                    {numbers.length
                      ? "No matching numbers"
                      : "No phone numbers to display"}
                  </p>
                  <p className="mt-2 text-muted-foreground">
                    {numbers.length
                      ? "Try another search or filter."
                      : "Existing numbers will appear here when account integration is connected."}
                  </p>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
      <p className="text-sm text-muted-foreground">
        UI preview. Number purchasing and provisioning are deferred.
      </p>
    </div>
  );
}
