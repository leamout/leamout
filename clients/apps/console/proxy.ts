import { type NextRequest, NextResponse } from "next/server";

export function proxy(request: NextRequest) {
  // Cookie presence allows rendering; Go still validates the session.
  if (!request.cookies.get("leamout-session")?.value) {
    const response = NextResponse.redirect(new URL("/log-in", request.url));

    response.headers.set("Cache-Control", "no-store");

    return response;
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/",
    "/agents/:path*",
    "/calls/:path*",
    "/phone-numbers/:path*",
    "/sip-trunks/:path*",
    "/ai-providers/:path*",
    "/tools/:path*",
    "/webhooks/:path*",
    "/recordings/:path*",
    "/settings/:path*",
    "/organizations/:path*",
    "/create-organization/:path*",
    "/create-account/password",
  ],
};
