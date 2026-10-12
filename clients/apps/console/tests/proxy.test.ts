import assert from "node:assert/strict";
import { test } from "node:test";
import { unstable_doesMiddlewareMatch } from "next/experimental/testing/server";
import { NextRequest } from "next/server";
import { config, proxy } from "../proxy";

test("protects console, onboarding, and password enrollment routes", () => {
  for (const url of [
    "/",
    "/agents",
    "/agents/new",
    "/agents/123",
    "/calls/123",
    "/phone-numbers",
    "/sip-trunks",
    "/ai-providers",
    "/tools/123",
    "/webhooks/123",
    "/recordings",
    "/settings/security",
    "/organizations",
    "/create-organization",
    "/create-account/password",
  ]) {
    assert.equal(unstable_doesMiddlewareMatch({ config, url }), true, url);
  }
});

test("leaves public auth pages, API rewrites, and assets accessible", () => {
  for (const url of [
    "/log-in",
    "/log-in/password",
    "/create-account",
    "/email-verification",
    "/forgot-password",
    "/reset-password",
    "/api/v1/users/me",
    "/api/v1/auth/start",
    "/_next/static/app.js",
    "/_next/image",
    "/favicon.ico",
  ]) {
    assert.equal(unstable_doesMiddlewareMatch({ config, url }), false, url);
  }
});

test("missing or empty session cookies redirect to login without caching", () => {
  for (const cookie of ["", "leamout-session=", "other=value"]) {
    const response = proxy(
      new NextRequest("https://console.example.com/agents", {
        headers: { cookie },
      }),
    );

    assert.equal(response.status, 307);
    assert.equal(
      response.headers.get("location"),
      "https://console.example.com/log-in",
    );
    assert.equal(response.headers.get("cache-control"), "no-store");
  }
});

test("cookie presence continues to the existing backend session check", () => {
  const response = proxy(
    new NextRequest("https://console.example.com/agents", {
      headers: { cookie: "leamout-session=opaque" },
    }),
  );

  assert.equal(response.headers.get("location"), null);
  assert.equal(response.headers.get("x-middleware-next"), "1");
});
