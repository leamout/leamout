"use client";

import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { useState } from "react";
import { z } from "zod";
import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

export type AgentProviderBinding = {
  role: "realtime" | "stt" | "llm" | "tts";
  provider: string;
  credentialId?: string;
};

export function AgentProviderSettings({
  bindings = [],
}: {
  bindings?: AgentProviderBinding[];
}) {
  const [engine, setEngine] = useState("realtime");
  const [notice, setNotice] = useState("");
  const roles = engine === "realtime" ? ["realtime"] : ["stt", "llm", "tts"];
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h2 className="text-lg font-semibold">Provider assignments</h2>
        <p className="text-sm text-muted-foreground">
          Use an organization credential or request a platform credential for
          each role. Provider availability and role compatibility have not been
          loaded.
        </p>
        <Link href="/ai-providers" className="text-sm hover:underline">
          Manage provider credentials
        </Link>
      </div>
      <div className="space-y-2">
        <label
          htmlFor="provider-engine-preview"
          className="block text-sm font-medium"
        >
          Preview engine
        </label>
        <select
          id="provider-engine-preview"
          className="h-9 rounded-md border bg-background px-3"
          value={engine}
          onChange={(event) => {
            setEngine(event.target.value);
            setNotice("");
          }}
        >
          <option value="realtime">Realtime</option>
          <option value="composable">STT → LLM → TTS</option>
        </select>
      </div>
      {roles.map((role) => {
        const binding = bindings.find((value) => value.role === role);
        return (
          <section
            key={role}
            className="space-y-4 rounded-lg border p-6"
            aria-label={`${role} provider assignment`}
          >
            <h3 className="font-semibold uppercase">{role}</h3>
            <p className="text-sm text-muted-foreground">
              {binding
                ? `Current provider: ${binding.provider} · ${binding.credentialId ? `Credential ${binding.credentialId}` : "Platform credential"}`
                : "Current assignment has not been loaded."}
            </p>
            <PreviewSettingsForm
              fields={[
                { name: "provider", label: "Provider ID", required: true },
                {
                  name: "credentialId",
                  label: "Organization credential ID (optional)",
                },
                {
                  name: "config",
                  label: "Provider configuration (JSON)",
                  type: "textarea",
                },
              ]}
              action="Preview provider assignment"
              validate={(values) => {
                const errors: { field: string; message: string }[] = [];
                if (
                  values.credentialId.trim() &&
                  !z.uuid().safeParse(values.credentialId.trim()).success
                )
                  errors.push({
                    field: "credentialId",
                    message: "Enter a valid credential UUID.",
                  });
                if (values.config.trim()) {
                  try {
                    const config = JSON.parse(values.config);
                    if (
                      config === null ||
                      Array.isArray(config) ||
                      typeof config !== "object"
                    )
                      throw new Error();
                  } catch {
                    errors.push({
                      field: "config",
                      message: "Enter a valid JSON object.",
                    });
                  }
                }
                return errors;
              }}
            />
            <p className="text-xs text-muted-foreground">
              Leave the credential ID blank to request platform credentials.
              Availability must be confirmed when integration is connected. Do
              not enter API secrets in configuration.
            </p>
            <Button
              variant="outline"
              disabled={!binding}
              onClick={() =>
                setNotice(
                  `Removal preview only. The ${role} assignment was not removed.`,
                )
              }
            >
              Preview assignment removal
            </Button>
          </section>
        );
      })}
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </div>
  );
}
