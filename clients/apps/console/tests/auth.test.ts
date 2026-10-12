import assert from "node:assert/strict";
import { once } from "node:events";
import { createServer } from "node:http";
import { after, before, test } from "node:test";
import { DELETE, GET, POST } from "../app/api/[...path]/route";
import { restoreFlow } from "../lib/auth/flow";

const transactionId = "6f4772e1-2e28-4b1e-8d14-31c9a4a905a1";
const flow = {
  transactionId,
  email: "user@example.com",
  intent: "recovery",
  stage: "reset",
  expiresAt: 2000,
  resendAt: 1000,
};

test("restores only non-secret flow fields and rejects expired or corrupt state", () => {
  assert.deepEqual(
    restoreFlow(
      JSON.stringify({ ...flow, password: "secret", code: "012345" }),
      1000,
    ),
    flow,
  );
  assert.equal(restoreFlow(JSON.stringify(flow), 2000), null);
  assert.equal(restoreFlow("{broken", 1000), null);
  assert.equal(
    restoreFlow(JSON.stringify({ ...flow, transactionId: "invalid" }), 1000),
    null,
  );
  assert.equal(restoreFlow(null), null);
});

const requests: {
  path: string;
  method: string;
  cookie?: string;
  body: string;
}[] = [];
const server = createServer(async (request, response) => {
  let body = "";
  for await (const chunk of request) body += chunk;
  requests.push({
    path: request.url ?? "",
    method: request.method ?? "",
    cookie: request.headers.cookie,
    body,
  });
  if (request.url === "/v1/sessions/current") {
    response.writeHead(204, {
      "Set-Cookie": "leamout_session=; Path=/; Max-Age=0; HttpOnly",
    });
    response.end();
    return;
  }
  if (request.url === "/v1/auth/otp/send") {
    response.writeHead(429, {
      "Content-Type": "application/json",
      "Retry-After": "60",
    });
    response.end(
      JSON.stringify({
        success: false,
        error: { code: "TOO_MANY_REQUESTS", message: "Wait before resending" },
      }),
    );
    return;
  }
  response.writeHead(200, {
    "Content-Type": "application/json",
    "Set-Cookie": "leamout_session=opaque; Path=/; HttpOnly; SameSite=Lax",
  });
  response.end(JSON.stringify({ success: true, data: { user_id: "user" } }));
});
const previousBackendURL = process.env.BACKEND_URL;
before(async () => {
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  const address = server.address();
  assert.ok(address && typeof address !== "string");
  process.env.BACKEND_URL = `http://127.0.0.1:${address.port}`;
});
after(async () => {
  if (previousBackendURL === undefined) delete process.env.BACKEND_URL;
  else process.env.BACKEND_URL = previousBackendURL;
  await new Promise<void>((resolve) => server.close(() => resolve()));
});
function context(path: string) {
  return { params: Promise.resolve({ path: path.split("/") }) };
}
function post(path: string, origin = "http://localhost:3001") {
  return new Request(`http://localhost:3001/api/${path}`, {
    method: "POST",
    headers: {
      origin,
      "Content-Type": "application/json",
      cookie: "leamout_session=opaque",
    },
    body: JSON.stringify({ transaction_id: transactionId, password: "secret" }),
  });
}

test("forwards JSON and cookies without caching the authentication response", async () => {
  const response = await POST(
    post("v1/auth/password/login"),
    context("v1/auth/password/login"),
  );
  assert.equal(response.status, 200);
  assert.equal(response.headers.get("cache-control"), "no-store");
  assert.match(response.headers.get("set-cookie") ?? "", /HttpOnly/);
  const request = requests.at(-1);
  assert.equal(request?.path, "/v1/auth/password/login");
  assert.equal(request?.cookie, "leamout_session=opaque");
  assert.equal(JSON.parse(request?.body ?? "{}").transaction_id, transactionId);
});
test("cross-origin writes and unlisted routes never reach the backend", async () => {
  const count = requests.length;
  assert.equal(
    (
      await POST(
        post("v1/auth/password/login", "https://attacker.example"),
        context("v1/auth/password/login"),
      )
    ).status,
    403,
  );
  assert.equal(
    (
      await GET(
        new Request("http://localhost:3001/api/metrics"),
        context("metrics"),
      )
    ).status,
    404,
  );
  assert.equal(requests.length, count);
});
test("rate-limit metadata is preserved", async () => {
  const response = await POST(
    post("v1/auth/otp/send"),
    context("v1/auth/otp/send"),
  );
  assert.equal(response.status, 429);
  assert.equal(response.headers.get("retry-after"), "60");
});
test("logout preserves a bodyless response and cookie deletion", async () => {
  const request = new Request("http://localhost:3001/api/v1/sessions/current", {
    method: "DELETE",
    headers: {
      origin: "http://localhost:3001",
      cookie: "leamout_session=opaque",
    },
  });
  const response = await DELETE(request, context("v1/sessions/current"));
  assert.equal(response.status, 204);
  assert.equal(await response.text(), "");
  assert.match(response.headers.get("set-cookie") ?? "", /Max-Age=0/);
});
