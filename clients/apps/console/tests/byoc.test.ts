import assert from "node:assert/strict";
import { test } from "node:test";
import { api } from "../lib/api";

test("BYOC trunk, endpoint, authentication, number and agent binding routes", async () => {
  const original = globalThis.fetch;
  const calls: { url: string; options: RequestInit }[] = [];

  globalThis.fetch = async (url, options) => {
    const target = String(url);
    calls.push({ url: target, options: options ?? {} });

    if (options?.method === "DELETE" || options?.method === "PUT") {
      return new Response(null, { status: 204 });
    }

    return Response.json({ success: true, data: {} });
  };

  try {
    const organizationId = "org-a";

    await api("/trunks/", { organizationId });
    await api("/trunks/", {
      method: "POST",
      organizationId,
      body: { name: "Carrier", direction: "inbound" },
    });
    await api("/trunks/trunk-a", {
      method: "PATCH",
      organizationId,
      body: { status: "active" },
    });
    await api("/trunks/trunk-a/validate", {
      method: "POST",
      organizationId,
    });
    await api("/trunks/trunk-a/endpoints", {
      method: "POST",
      organizationId,
      body: { host: "sip.example.com", port: 5060, transport: "udp" },
    });
    await api("/trunks/trunk-a/inbound-auth", {
      method: "PUT",
      organizationId,
      body: { method: "ip" },
    });
    await api("/trunks/trunk-a/outbound-auth", {
      method: "PUT",
      organizationId,
      body: { method: "none" },
    });
    await api("/trunks/trunk-a/source-ips", {
      method: "POST",
      organizationId,
      body: { cidr: "198.51.100.0/24" },
    });
    await api("/numbers/", {
      method: "POST",
      organizationId,
      idempotencyKey: "number-a",
      body: { number: "+233201234567", country_code: "GH" },
    });
    await api("/numbers/number-a/trunk", {
      method: "PATCH",
      organizationId,
      body: { trunk_id: "trunk-a" },
    });
    await api("/voice-agents/agent-a/bindings", {
      method: "POST",
      organizationId,
      body: { phone_number_id: "number-a" },
    });
    await api("/voice-agents/agent-a/bindings/binding-a", {
      method: "DELETE",
      organizationId,
    });

    assert.deepEqual(
      calls.map((call) => [call.options.method, call.url]),
      [
        ["GET", "/api/v1/trunks/"],
        ["POST", "/api/v1/trunks/"],
        ["PATCH", "/api/v1/trunks/trunk-a"],
        ["POST", "/api/v1/trunks/trunk-a/validate"],
        ["POST", "/api/v1/trunks/trunk-a/endpoints"],
        ["PUT", "/api/v1/trunks/trunk-a/inbound-auth"],
        ["PUT", "/api/v1/trunks/trunk-a/outbound-auth"],
        ["POST", "/api/v1/trunks/trunk-a/source-ips"],
        ["POST", "/api/v1/numbers/"],
        ["PATCH", "/api/v1/numbers/number-a/trunk"],
        ["POST", "/api/v1/voice-agents/agent-a/bindings"],
        ["DELETE", "/api/v1/voice-agents/agent-a/bindings/binding-a"],
      ],
    );

    for (const call of calls) {
      const headers = new Headers(call.options.headers);
      assert.equal(headers.get("X-Organization-ID"), organizationId);
      assert.equal(headers.get("Authorization"), null);
      assert.equal(call.options.credentials, "same-origin");
    }

    assert.equal(
      new Headers(calls[8].options.headers).get("Idempotency-Key"),
      "number-a",
    );
  } finally {
    globalThis.fetch = original;
  }
});
