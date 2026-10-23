import assert from "node:assert/strict";
import { test } from "node:test";
import { api } from "../lib/api";

test("AI integrations and Agent provider operations use organization-scoped endpoints", async () => {
  const originalFetch = globalThis.fetch;
  const calls: { url: string; options: RequestInit }[] = [];
  globalThis.fetch = async (url, options) => {
    calls.push({ url: String(url), options: options ?? {} });
    if (options?.method === "DELETE") {
      return new Response(null, { status: 204 });
    }
    return Response.json({ success: true, data: {} });
  };

  try {
    const organizationId = "org-test";
    await api("/ai-integrations/", { organizationId });
    await api("/ai-integrations/", {
      method: "POST",
      organizationId,
      body: { provider: "openai", name: "production", secret: "test-key" },
    });
    await api("/ai-integrations/credential-a/verify", {
      method: "POST",
      organizationId,
    });
    await api("/ai-integrations/credential-a/rotate", {
      method: "POST",
      organizationId,
      body: { secret: "replacement-key" },
    });
    await api("/ai-integrations/credential-a", {
      method: "DELETE",
      organizationId,
    });
    await api("/voice-agents/agent-a/providers/", { organizationId });
    await api("/voice-agents/agent-a/providers/realtime", {
      method: "PUT",
      organizationId,
      body: {
        provider: "openai",
        credential_id: "credential-a",
        config: {},
      },
    });
    await api("/voice-agents/agent-a/providers/realtime", {
      method: "DELETE",
      organizationId,
    });
    await api("/voice-agents/agent-a/readiness", { organizationId });

    assert.deepEqual(
      calls.map(({ url, options }) => [options.method, url]),
      [
        ["GET", "/api/v1/ai-integrations/"],
        ["POST", "/api/v1/ai-integrations/"],
        ["POST", "/api/v1/ai-integrations/credential-a/verify"],
        ["POST", "/api/v1/ai-integrations/credential-a/rotate"],
        ["DELETE", "/api/v1/ai-integrations/credential-a"],
        ["GET", "/api/v1/voice-agents/agent-a/providers/"],
        ["PUT", "/api/v1/voice-agents/agent-a/providers/realtime"],
        ["DELETE", "/api/v1/voice-agents/agent-a/providers/realtime"],
        ["GET", "/api/v1/voice-agents/agent-a/readiness"],
      ],
    );

    for (const { options } of calls) {
      const headers = new Headers(options.headers);
      assert.equal(headers.get("X-Organization-ID"), organizationId);
      assert.equal(headers.get("Authorization"), null);
      assert.equal(options.credentials, "same-origin");
      assert.equal(options.cache, "no-store");
    }
    assert.equal(
      JSON.parse(String(calls[6].options.body)).credential_id,
      "credential-a",
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});
