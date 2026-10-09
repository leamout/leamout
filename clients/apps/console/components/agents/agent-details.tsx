"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@leamout/ui/components/tabs";
import Link from "next/link";
import { AgentForm } from "@/components/agents/agent-form";
import { AgentSettingsForm } from "@/components/agents/agent-settings-form";

const sections = [
  {
    value: "models",
    title: "Models & voice",
    description:
      "Configure a realtime model or an STT → LLM → TTS pipeline, provider credentials, and voice.",
  },
  {
    value: "tools",
    title: "Tools",
    description:
      "Assign tools and configure their availability during conversations.",
    href: "/tools",
    link: "Manage tools",
  },
  {
    value: "calls",
    title: "Call settings",
    description:
      "Configure duration limits, silence handling, interruptions, recording, and transfer behavior.",
    href: "/phone-numbers",
    link: "Manage phone numbers",
  },
  {
    value: "test",
    title: "Test",
    description:
      "Interactive agent testing will be available when runtime integration is connected.",
  },
  {
    value: "activity",
    title: "Activity",
    description:
      "This agent’s calls will appear here. Conversation history and diagnostics belong in call details.",
    href: "/calls",
    link: "View calls",
  },
];

export function AgentDetails({ agentId }: { agentId: string }) {
  return (
    <div className="space-y-6">
      <Link
        href="/agents"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to agents
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">Agent details</h1>
        <p className="break-all text-sm text-muted-foreground">
          Agent ID: {agentId}
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview. Agent data has not been loaded.
        </p>
      </div>
      <Tabs defaultValue="configuration" className="gap-6">
        <div className="overflow-x-auto">
          <TabsList aria-label="Agent settings">
            <TabsTrigger value="configuration">Configuration</TabsTrigger>
            {sections.map((section) => (
              <TabsTrigger key={section.value} value={section.value}>
                {section.title}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
        <TabsContent value="configuration">
          <AgentForm />
        </TabsContent>
        {sections.map((section) => (
          <TabsContent key={section.value} value={section.value}>
            {section.value === "models" || section.value === "calls" ? (
              <AgentSettingsForm section={section.value} />
            ) : (
              <div className="space-y-4 rounded-lg border p-6">
                <h2 className="text-lg font-semibold">{section.title}</h2>
                <p className="max-w-2xl text-muted-foreground">
                  {section.description}
                </p>
                {section.href && (
                  <Button
                    variant="outline"
                    render={<Link href={section.href} />}
                  >
                    {section.link}
                  </Button>
                )}
                <p className="text-sm text-muted-foreground">
                  This section is being prepared.
                </p>
              </div>
            )}
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}
