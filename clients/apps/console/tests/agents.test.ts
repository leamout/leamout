import assert from "node:assert/strict";
import { test } from "node:test";
import {
  createVoiceAgent,
  disableVoiceAgent,
  getVoiceAgent,
  listVoiceAgents,
  updateVoiceAgent,
  voiceAgentKeys,
} from "../components/agents/api";
import type { VoiceAgentInput } from "../components/agents/model";

const input: VoiceAgentInput = {
  name: "Support",
  engine: "realtime",
  instructions: "Handle support calls.",
  language: "en",
  engine_config: {},
  interruption_policy: "allow",
  recording_policy: "none",
};

test("Voice Agent operations preserve organization scope and backend methods", async () => {
  const originalFetch = globalThis.fetch;
  const requests: Array<{ url: string; options: RequestInit }> = [];

  globalThis.fetch = async (path, options) => {
    requests.push({ url: String(path), options: options ?? {} });
    const method = options?.method ?? "GET";
    if (method === "DELETE") return new Response(null, { status: 204 });
    return Response.json({ success: true, data: method === "GET" ? { voice_agents: [] } : { id: "agent-a" } });
  };

  try {
    await listVoiceAgents("org-a");
    await getVoiceAgent("org-a", "agent-a");
    await createVoiceAgent("org-a", input, "create-a");
    await updateVoiceAgent("org-a", "agent-a", { name: "Updated" });
    await disableVoiceAgent("org-a", "agent-a");

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
    }
    assert.equal(
      new Headers(requests[2].options.headers).get("Idempotency-Key"),
      "create-a",
    );
    assert.deepEqual(JSON.parse(String(requests[2].options.body)), input);
    assert.deepEqual(JSON.parse(String(requests[3].options.body)), {
      name: "Updated",
    });
    assert.deepEqual(voiceAgentKeys.list("org-a"), ["voice-agents", "org-a"]);
    assert.deepEqual(voiceAgentKeys.detail("org-b", "agent-a"), [
      "voice-agents", "org-b", "agent-a",
    ]);
  } finally {
    globalThis.fetch = originalFetch;
  }
});
