import assert from "node:assert/strict";
import { test } from "node:test";
import { api } from "../lib/api";

test("Tools and webhooks use existing organization-scoped backend routes", async () => {
  const original = globalThis.fetch;
  const calls: { path: string; options: RequestInit }[] = [];
  globalThis.fetch = async (path, options) => {
    calls.push({ path: String(path), options: options ?? {} });
    if (options?.method === "DELETE") {
      return new Response(null, { status: 204 });
    }
    return Response.json({ success: true, data: {} });
  };
  try {
    const organizationId = "org-a";
    const toolPath = "/voice-agents/agent-a/tools/";
    await api(toolPath, { organizationId });
    await api(toolPath, {
      method: "POST",
      organizationId,
      body: {
        type: "builtin",
        name: "hold_call",
        description: "Hold call",
        parameters: {},
      },
    });
    await api(toolPath + "tool-a", {
      method: "PATCH",
      organizationId, body: { enabled: false },
    });
    await api(toolPath + "tool-a/rotate-signing-secret", {
      method: "POST",
      organizationId,
    });
    await api(toolPath + "tool-a", { method: "DELETE",
      organizationId });
    await api("/webhooks/", { organizationId });
    await api("/webhooks/", {
      method: "POST",
      organizationId,
      body: {
        url: "https://example.com/hooks",
        subscribed_events: ["call.completed"],
        enabled: true,
      },
    });
    await api("/webhooks/webhook-a", { organizationId });
    await api("/webhooks/webhook-a", {
      method: "PATCH",
      organizationId, body: { enabled: false },
    });
    await api("/webhooks/webhook-a/rotate-secret", {
      method: "POST",
      organizationId,
    });
    await api("/webhooks/webhook-a/test", { method: "POST", organizationId });
    await api("/webhooks/webhook-a/deliveries?limit=50&offset=0", { organizationId });
    await api("/webhooks/webhook-a/deliveries/delivery-a/retry", {
      method: "POST",
      organizationId,
    });
    await api("/webhooks/webhook-a", { method: "DELETE",
      organizationId });

    assert.deepEqual(
      calls.map(({ path, options }) => [options.method, path]),
      [
        ["GET", "/api/v1/voice-agents/agent-a/tools/"],
        ["POST", "/api/v1/voice-agents/agent-a/tools/"],
        ["PATCH", "/api/v1/voice-agents/agent-a/tools/tool-a"],
        [
          "POST",
          "/api/v1/voice-agents/agent-a/tools/tool-a/rotate-signing-secret",
        ],
        ["DELETE", "/api/v1/voice-agents/agent-a/tools/tool-a"],
        ["GET", "/api/v1/webhooks/"],
        ["POST", "/api/v1/webhooks/"],
        ["GET", "/api/v1/webhooks/webhook-a"],
        ["PATCH", "/api/v1/webhooks/webhook-a"],
        ["POST", "/api/v1/webhooks/webhook-a/rotate-secret"],
        ["POST", "/api/v1/webhooks/webhook-a/test"],
        ["GET", "/api/v1/webhooks/webhook-a/deliveries?limit=50&offset=0"],
        [
          "POST",
          "/api/v1/webhooks/webhook-a/deliveries/delivery-a/retry",
        ],
        ["DELETE", "/api/v1/webhooks/webhook-a"],
      ],
    );
    for (const { options } of calls) {
      const headers = new Headers(options.headers);
      assert.equal(headers.get("X-Organization-ID"), "org-a");
      assert.equal(headers.get("Authorization"), null);
      assert.equal(options.credentials, "same-origin");
      assert.equal(options.cache, "no-store");
    }
  } finally {
    globalThis.fetch = original;
  }
});
