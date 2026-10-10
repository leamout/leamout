"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldLabel,
} from "@leamout/ui/components/field";
import Link from "next/link";
import { useState } from "react";
import type { ResourceOption } from "@/components/phone-numbers/types";

export function PhoneNumberRoutingForm({
  agents = [],
  assignedAgent,
  loaded = false,
}: {
  agents?: ResourceOption[];
  assignedAgent?: ResourceOption;
  loaded?: boolean;
}) {
  const [agentId, setAgentId] = useState(assignedAgent?.id ?? "");
  const [notice, setNotice] = useState("");
  const options =
    assignedAgent && !agents.some((agent) => agent.id === assignedAgent.id)
      ? [assignedAgent, ...agents]
      : agents;
  return (
    <form
      className="max-w-2xl space-y-6"
      onSubmit={(event) => {
        event.preventDefault();
        setNotice("Preview only. Agent routing was not saved.");
      }}
    >
      <h2 className="text-lg font-semibold">Inbound agent routing</h2>
      <Field>
        <FieldLabel htmlFor="number-agent">Assigned agent</FieldLabel>
        <select
          id="number-agent"
          className="h-9 rounded-md border bg-background px-3"
          value={agentId}
          onChange={(event) => {
            setAgentId(event.target.value);
            setNotice("");
          }}
        >
          <option value="">Unassigned</option>
          {options.map((agent) => (
            <option key={agent.id} value={agent.id}>
              {agent.name}
            </option>
          ))}
        </select>
        <FieldDescription>
          {options.length
            ? "Choose the agent to receive calls to this number."
            : "Agent options have not been loaded."}{" "}
          <Link href="/agents" className="underline">
            Manage agents
          </Link>
        </FieldDescription>
      </Field>
      <Button type="submit" disabled={!loaded}>
        Save routing
      </Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
