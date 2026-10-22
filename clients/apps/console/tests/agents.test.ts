import assert from "node:assert/strict";
import { test } from "node:test";
import type { VoiceAgent, VoiceAgentInput } from "../components/agents/types";
import { api } from "../lib/api";

const input: VoiceAgentInput = {
  name: "Support",
  engine: "realtime",
  instructions: "Handle support calls.",
  language: "en",
  engine_config: {},
  interruption_policy: "allow",
  recording_policy: "none",
};

const agent: VoiceAgent = {
  id: "agent-a",
  organization_id: "org-a",
  ...input,
  language: "en",
  status: "active",
  configuration_revision: 1,
  created_at: "2026-10-22T00:00:00Z",
  updated_at: "2026-10-22T00:00:00Z",
};

test("Voice Agent requests use the shared API directly with organization scope", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: string; options: RequestInit }> = [];

  globalThis.fetch = async (path, options) => {
    const url = String(path);
    requests.push({ url, options: options ?? {} });
    if (options?.method === "DELETE") {
      return new Response(null, { status: 204 });
    }
    const data = url.endsWith("/voice-agents/")
      ? options?.method === "GET"
        ? { voice_agents: [agent] }
        : agent
      : agent;
    return Response.json({ success: true, data });
  };

  try {
    const listed = await api<{ voice_agents: VoiceAgent[] }>("/voice-agents/", {
      organizationId: "org-a",
    });
    const fetched = await api<VoiceAgent>("/voice-agents/agent-a", {
      organizationId: "org-a",
    });
    const created = await api<VoiceAgent>("/voice-agents/", {
      method: "POST",
      organizationId: "org-a",
      idempotencyKey: "create-a",
      body: input,
    });
    const updated = await api<VoiceAgent>("/voice-agents/agent-a", {
      method: "PATCH",
      organizationId: "org-a",
      body: { name: "Updated" },
    });
    const disabled = await api<void>("/voice-agents/agent-a", {
      method: "DELETE",
      organizationId: "org-a",
    });
    assert.equal(listed.voice_agents[0].id, "agent-a");
    assert.equal(fetched.id, "agent-a");
    assert.equal(created.id, "agent-a");
    assert.equal(updated.id, "agent-a");
    assert.equal(disabled, undefined);

    assert.deepEqual(
      requests.map(({ url, options }) => [options.method, url]),
      [
        ["GET", "/api/v1/voice-agents/"],
        ["GET", "/api/v1/voice-agents/agent-a"],
        ["POST", "/api/v1/voice-agents/"],
        ["PATCH", "/api/v1/voice-agents/agent-a"],
        ["DELETE", "/api/v1/voice-agents/agent-a"],
      ],
    );
    for (const { options } of requests) {
      const headers = new Headers(options.headers);
      assert.equal(headers.get("X-Organization-ID"), "org-a");
      assert.equal(headers.get("Authorization"), null);
      assert.equal(options.credentials, "same-origin");
      assert.equal(options.cache, "no-store");
    }
    assert.equal(
      new Headers(requests[2].options.headers).get("Idempotency-Key"),
      "create-a",
    );
    assert.deepEqual(JSON.parse(String(requests[2].options.body)), input);
    assert.deepEqual(JSON.parse(String(requests[3].options.body)), {
      name: "Updated",
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});
