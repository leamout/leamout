"use client";

import { Button } from "@leamout/ui/components/button";
import {
  Field,
  FieldDescription,
  FieldLabel,
} from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useState } from "react";

export function AgentSettingsForm({
  section,
}: {
  section: "models" | "calls";
}) {
  const [mode, setMode] = useState("realtime");
  const [notice, setNotice] = useState("");
  const fields =
    section === "models"
      ? mode === "realtime"
        ? ["Realtime provider", "Realtime model", "Voice"]
        : [
            "STT provider",
            "STT model",
            "LLM provider",
            "LLM model",
            "TTS provider",
            "TTS model",
            "Voice",
          ]
      : [
          "Maximum call duration (seconds)",
          "Silence timeout (seconds)",
          "Transfer destination",
        ];

  return (
    <form
      className="max-w-2xl space-y-6"
      onChange={() => setNotice("")}
      onSubmit={(event) => {
        event.preventDefault();
        setNotice("Preview only. Changes were not saved.");
      }}
    >
      <h2 className="text-lg font-semibold">
        {section === "models" ? "Models & voice" : "Call settings"}
      </h2>
      {section === "models" && (
        <Field>
          <FieldLabel htmlFor="agent-mode">Conversation mode</FieldLabel>
          <select
            id="agent-mode"
            className="h-9 rounded-md border bg-background px-3"
            value={mode}
            onChange={(event) => setMode(event.target.value)}
          >
            <option value="realtime">Realtime</option>
            <option value="pipeline">STT → LLM → TTS</option>
          </select>
          <FieldDescription>
            Provider and model availability will be connected later.
          </FieldDescription>
        </Field>
      )}
      {fields.map((label) => (
        <Field key={`${mode}-${label}`}>
          <FieldLabel htmlFor={`agent-${section}-${label}`}>{label}</FieldLabel>
          <Input
            id={`agent-${section}-${label}`}
            type={label.includes("(seconds)") ? "number" : "text"}
            min={label.includes("(seconds)") ? 1 : undefined}
            step={label.includes("(seconds)") ? 1 : undefined}
          />
        </Field>
      ))}
      {section === "calls" && (
        <div className="space-y-3">
          <label className="flex items-center gap-3">
            <input type="checkbox" />
            Allow interruptions
          </label>
          <label className="flex items-center gap-3">
            <input type="checkbox" />
            Record calls
          </label>
        </div>
      )}
      <Button type="submit">Save changes</Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
    </form>
  );
}
