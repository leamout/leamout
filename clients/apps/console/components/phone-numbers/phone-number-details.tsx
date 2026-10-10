"use client";

import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import Link from "next/link";
import { PhoneNumberForm } from "@/components/phone-numbers/phone-number-form";
import { PhoneNumberRoutingForm } from "@/components/phone-numbers/phone-number-routing-form";
import type {
  PhoneNumber,
  ResourceOption,
} from "@/components/phone-numbers/types";

export function PhoneNumberDetails({
  numberId,
  number,
  trunks = [],
  agents = [],
}: {
  numberId: string;
  number?: PhoneNumber;
  trunks?: ResourceOption[];
  agents?: ResourceOption[];
}) {
  return (
    <div className="space-y-6">
      <Link
        href="/phone-numbers"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to phone numbers
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          {number?.number ?? "Phone number details"}
        </h1>
        <p className="break-all text-sm text-muted-foreground">
          Number ID: {numberId}
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview.{" "}
          {number
            ? "Preview configuration is shown below."
            : "Number configuration has not been loaded."}
        </p>
      </div>
      <dl className="grid gap-6 rounded-lg border p-6 sm:grid-cols-3">
        {[
          ["Country", number?.countryCode ?? "—"],
          ["Status", number?.status ?? "Unavailable"],
          [
            "Voice",
            number
              ? number.voiceEnabled
                ? "Enabled"
                : "Disabled"
              : "Unavailable",
          ],
        ].map(([label, value]) => (
          <div key={label}>
            <dt className="text-sm text-muted-foreground">{label}</dt>
            <dd className="mt-1 font-medium">{value}</dd>
          </div>
        ))}
      </dl>
      <Tabs defaultValue="configuration" className="gap-6">
        <TabsList aria-label="Phone number settings">
          <TabsTrigger value="configuration">Configuration</TabsTrigger>
          <TabsTrigger value="routing">Routing</TabsTrigger>
        </TabsList>
        <TabsContent value="configuration">
          <PhoneNumberForm number={number} trunks={trunks} />
        </TabsContent>
        <TabsContent value="routing">
          <PhoneNumberRoutingForm
            agents={agents}
            assignedAgent={number?.agent}
            loaded={Boolean(number)}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
