"use client";

import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import Link from "next/link";
import { TrunkAuthForm } from "@/components/sip-trunks/trunk-auth-form";
import { TrunkConnectionForm } from "@/components/sip-trunks/trunk-connection-form";
import { TrunkForm } from "@/components/sip-trunks/trunk-form";

export function TrunkDetails({ trunkId }: { trunkId: string }) {
  return (
    <div className="space-y-6">
      <Link
        href="/sip-trunks"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to SIP trunks
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          SIP trunk details
        </h1>
        <p className="break-all text-sm text-muted-foreground">
          Trunk ID: {trunkId}
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview. Trunk configuration has not been loaded.
        </p>
      </div>
      <Tabs defaultValue="configuration" className="gap-6">
        <div className="overflow-x-auto">
          <TabsList aria-label="SIP trunk settings">
            {["Configuration", "Endpoints", "Authentication", "Routing"].map(
              (label) => (
                <TabsTrigger key={label} value={label.toLowerCase()}>
                  {label}
                </TabsTrigger>
              ),
            )}
          </TabsList>
        </div>
        <TabsContent value="configuration">
          <TrunkForm />
        </TabsContent>
        <TabsContent value="endpoints">
          <TrunkConnectionForm />
        </TabsContent>
        <TabsContent value="authentication">
          <div className="space-y-10">
            <TrunkAuthForm inbound />
            <TrunkAuthForm inbound={false} />
          </div>
        </TabsContent>
        <TabsContent value="routing">
          <div className="space-y-4 rounded-lg border p-6">
            <h2 className="text-lg font-semibold">Phone number routing</h2>
            <p className="text-muted-foreground">
              Assigned numbers and routing rules have not been loaded. Configure
              agent assignments from phone number details.
            </p>
            <Link href="/phone-numbers" className="text-sm underline">
              Manage phone numbers
            </Link>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}
