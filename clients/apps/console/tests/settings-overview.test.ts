import assert from "node:assert/strict";
import { test } from "node:test";
import { api } from "../lib/api";

test("Overview and Settings use the existing organization-scoped routes", async () => {
  const originalFetch = globalThis.fetch;
  const requests: { url: string; options: RequestInit }[] = [];
  globalThis.fetch = async (url, options) => {
    requests.push({ url: String(url), options: options ?? {} });
    if (options?.method === "DELETE") {
      return new Response(null, { status: 204 });
    }
    return Response.json({ success: true, data: {} });
  };

  try {
    const organizationId = "org-a";
    await api("/voice-agents/", { organizationId });
    await api("/trunks/", { organizationId });
    await api("/numbers/", { organizationId });
    await api("/ai-integrations/", { organizationId });
    await api("/calls/?limit=10&offset=0", { organizationId });
    await api("/organizations/org-a", { organizationId });
    await api("/organizations/org-a", {
      method: "PATCH",
      organizationId,
      body: { name: "Updated" },
    });
    await api("/organizations/org-a/members/", { organizationId });
    await api("/organizations/org-a/members/member-a", {
      method: "PATCH",
      organizationId,
      body: { role: "admin" },
    });
    await api("/organizations/org-a/members/member-a", {
      method: "DELETE",
      organizationId,
    });
    await api("/organizations/org-a/invitations/", { organizationId });
    await api("/organizations/org-a/invitations/", {
      method: "POST",
      organizationId,
      body: { email: "user@example.com", role: "member" },
    });
    await api("/organizations/org-a/invitations/invite-a", {
      method: "DELETE",
      organizationId,
    });
    await api("/audit-events/?limit=50&offset=0", { organizationId });

    assert.deepEqual(
      requests.map(({ url, options }) => [options.method, url]),
      [
        ["GET", "/api/v1/voice-agents/"],
        ["GET", "/api/v1/trunks/"],
        ["GET", "/api/v1/numbers/"],
        ["GET", "/api/v1/ai-integrations/"],
        ["GET", "/api/v1/calls/?limit=10&offset=0"],
        ["GET", "/api/v1/organizations/org-a"],
        ["PATCH", "/api/v1/organizations/org-a"],
        ["GET", "/api/v1/organizations/org-a/members/"],
        ["PATCH", "/api/v1/organizations/org-a/members/member-a"],
        ["DELETE", "/api/v1/organizations/org-a/members/member-a"],
        ["GET", "/api/v1/organizations/org-a/invitations/"],
        ["POST", "/api/v1/organizations/org-a/invitations/"],
        ["DELETE", "/api/v1/organizations/org-a/invitations/invite-a"],
        ["GET", "/api/v1/audit-events/?limit=50&offset=0"],
      ],
    );

    for (const { options } of requests) {
      const headers = new Headers(options.headers);
      assert.equal(headers.get("X-Organization-ID"), organizationId);
      assert.equal(headers.get("Authorization"), null);
      assert.equal(options.credentials, "same-origin");
      assert.equal(options.cache, "no-store");
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});
