// Only console authentication endpoints are forwarded in this first integration.
const routes = new Set([
  "POST v1/auth/start",
  "POST v1/auth/password/login",
  "POST v1/auth/otp/send",
  "POST v1/auth/otp/verify",
  "POST v1/auth/password/enroll",
  "POST v1/auth/password/recovery",
  "POST v1/auth/password/reset",
  "GET v1/users/me",
  "DELETE v1/sessions/current",
]);

async function proxy(
  request: Request,
  context: { params: Promise<{ path: string[] }> },
) {
  const { path } = await context.params;
  const route = path.join("/");
  if (!routes.has(`${request.method} ${route}`))
    return new Response(null, { status: 404 });

  // Reject cross-origin writes, including login requests, before forwarding cookies.
  if (
    request.method !== "GET" &&
    request.headers.get("origin") !== new URL(request.url).origin
  ) {
    return Response.json(
      {
        success: false,
        error: { code: "FORBIDDEN", message: "Invalid request origin." },
      },
      { status: 403 },
    );
  }

  if (
    request.method === "POST" &&
    !request.headers.get("content-type")?.startsWith("application/json")
  ) {
    return new Response(null, { status: 415 });
  }

  const headers = new Headers({ Accept: "application/json" });
  for (const name of ["cookie", "content-type", "user-agent"]) {
    const value = request.headers.get(name);
    if (value) headers.set(name, value);
  }

  try {
    const upstream = await fetch(
      `${(process.env.BACKEND_URL ?? "http://localhost:8080").replace(/\/$/, "")}/${route}`,
      {
        method: request.method,
        headers,
        body: request.method === "POST" ? await request.text() : undefined,
        cache: "no-store",
        redirect: "manual",
        signal: AbortSignal.any([request.signal, AbortSignal.timeout(20_000)]),
      },
    );
    const responseHeaders = new Headers({ "Cache-Control": "no-store" });
    for (const name of ["content-type", "retry-after"]) {
      const value = upstream.headers.get(name);
      if (value) responseHeaders.set(name, value);
    }
    for (const cookie of upstream.headers.getSetCookie())
      responseHeaders.append("Set-Cookie", cookie);
    return new Response(upstream.body, {
      status: upstream.status,
      headers: responseHeaders,
    });
  } catch {
    return Response.json(
      {
        success: false,
        error: {
          code: "UPSTREAM_UNAVAILABLE",
          message: "Leamout is temporarily unavailable. Please try again.",
        },
      },
      { status: 502 },
    );
  }
}

export { proxy as GET, proxy as POST, proxy as DELETE };
