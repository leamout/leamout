"use client";

import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { useState } from "react";
import { z } from "zod";
import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

export type AgentRoutingBinding = {
  id: string;
  phoneNumberId: string;
  phoneNumber?: string;
};

export function AgentRouting({
  bindings = [],
}: {
  bindings?: AgentRoutingBinding[];
}) {
  const [notice, setNotice] = useState("");
  const [removing, setRemoving] = useState<AgentRoutingBinding>();
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold">Phone number routing</h2>
        <p className="text-sm text-muted-foreground">
          Bind existing BYOC phone numbers to this agent.
        </p>
        <Link href="/phone-numbers" className="text-sm hover:underline">
          Manage phone numbers
        </Link>
      </div>
      {bindings.length ? (
        <ul className="space-y-3" aria-label="Agent phone number bindings">
          {bindings.map((binding) => (
            <li
              key={binding.id}
              className="flex flex-wrap items-center justify-between gap-4 rounded-lg border p-4"
            >
              <Link
                className="break-all font-medium hover:underline"
                href={`/phone-numbers/${encodeURIComponent(binding.phoneNumberId)}`}
              >
                {binding.phoneNumber ?? binding.phoneNumberId}
              </Link>
              <Button
                variant="outline"
                onClick={() => {
                  setRemoving(binding);
                  setNotice("");
                }}
              >
                Remove binding
              </Button>
            </li>
          ))}
        </ul>
      ) : (
        <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
          No phone number bindings to display. Routing data is not connected.
        </p>
      )}
      {removing && (
        <div className="space-y-3 rounded-lg border p-6">
          <p className="font-medium">
            Remove binding for {removing.phoneNumber ?? removing.phoneNumberId}?
          </p>
          <p className="text-sm text-muted-foreground">
            The phone number itself remains in your organization.
          </p>
          <div className="flex gap-3">
            <Button
              variant="destructive"
              onClick={() => {
                setRemoving(undefined);
                setNotice("Removal preview only. No binding was removed.");
              }}
            >
              Confirm removal preview
            </Button>
            <Button variant="outline" onClick={() => setRemoving(undefined)}>
              Cancel
            </Button>
          </div>
        </div>
      )}
      <section
        className="space-y-4 rounded-lg border p-6"
        aria-labelledby="agent-bind-number-title"
      >
        <h3 id="agent-bind-number-title" className="font-semibold">
          Bind a phone number
        </h3>
        <PreviewSettingsForm
          fields={[
            {
              name: "phoneNumberId",
              label: "Existing phone number ID",
              required: true,
            },
          ]}
          action="Preview number binding"
          validate={(values) =>
            z.uuid().safeParse(values.phoneNumberId.trim()).success
              ? []
              : [
                  {
                    field: "phoneNumberId",
                    message: "Enter a valid phone number UUID.",
                  },
                ]
          }
        />
      </section>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </div>
  );
}
