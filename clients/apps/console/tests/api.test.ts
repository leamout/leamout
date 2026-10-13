import assert from "node:assert/strict";
import { test } from "node:test";
import { ApiError, api } from "../lib/api";

test("checkout requests include organization scope, idempotency, and JSON body", async () => {
  const originalFetch = globalThis.fetch;
  const calls: {
    path: string;
    options: RequestInit | undefined;
  }[] = [];

  globalThis.fetch = async (path, options) => {
    calls.push({
      path: String(path),
      options,
    });
    return Response.json({
      success: true,
      data: {
        checkout_url: "https://checkout.polar.sh/session",
      },
    });
  };

  try {
    const result = await api<{ checkout_url: string }>("/billing/checkout", {
      method: "POST",
      organizationId: "org-first",
      idempotencyKey: "checkout-first",
      body: {
        plan_code: "developer",
      },
    });
    const request = calls[0];
    const headers = new Headers(request.options?.headers);

    assert.equal(result.checkout_url, "https://checkout.polar.sh/session");
    assert.equal(request.path, "/api/v1/billing/checkout");
    assert.equal(request.options?.method, "POST");
    assert.equal(request.options?.credentials, "same-origin");
    assert.equal(request.options?.cache, "no-store");
    assert.equal(headers.get("X-Organization-ID"), "org-first");
    assert.equal(headers.get("Idempotency-Key"), "checkout-first");
    assert.equal(headers.get("Content-Type"), "application/json");
    assert.equal(headers.get("Authorization"), null);
    assert.deepEqual(JSON.parse(String(request.options?.body)), {
      plan_code: "developer",
    });
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("requests do not share organization, idempotency, or authorization headers", async () => {
  const originalFetch = globalThis.fetch;
  const requests: RequestInit[] = [];

  globalThis.fetch = async (_path, options) => {
    requests.push(options ?? {});
    return Response.json({
      success: true,
      data: {},
    });
  };

  try {
    await Promise.all([
      api("/billing/checkout", {
        method: "POST",
        organizationId: "org-first",
        idempotencyKey: "checkout-first",
        body: {
          plan_code: "pro",
        },
      }),
      api("/billing/status", {
        organizationId: "org-second",
      }),
      api("/users/me"),
    ]);

    const first = new Headers(requests[0].headers);
    const second = new Headers(requests[1].headers);
    const third = new Headers(requests[2].headers);

    assert.notEqual(requests[0].headers, requests[1].headers);
    assert.equal(first.get("X-Organization-ID"), "org-first");
    assert.equal(first.get("Idempotency-Key"), "checkout-first");
    assert.equal(second.get("X-Organization-ID"), "org-second");
    assert.equal(second.get("Idempotency-Key"), null);
    assert.equal(second.get("Content-Type"), null);
    assert.equal(third.get("X-Organization-ID"), null);

    for (const request of requests) {
      assert.equal(new Headers(request.headers).get("Authorization"), null);
      assert.equal(request.credentials, "same-origin");
    }
  } finally {
    globalThis.fetch = originalFetch;
  }
});

test("forbidden billing responses reject instead of returning successful data", async () => {
  const originalFetch = globalThis.fetch;

  globalThis.fetch = async () =>
    Response.json(
      {
        success: false,
        error: {
          code: "FORBIDDEN",
          message: "Only organization administrators can manage billing.",
        },
      },
      {
        status: 403,
      },
    );

  try {
    await assert.rejects(
      api("/billing/checkout", {
        method: "POST",
        organizationId: "org-member",
        body: {
          plan_code: "scale",
        },
      }),
      (error: unknown) => {
        assert.ok(error instanceof ApiError);
        assert.equal(error.status, 403);
        assert.equal(error.code, "FORBIDDEN");
        assert.equal(
          error.message,
          "Only organization administrators can manage billing.",
        );
        return true;
      },
    );
  } finally {
    globalThis.fetch = originalFetch;
  }
});
