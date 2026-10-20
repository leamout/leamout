import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  transpilePackages: ["@leamout/ui"],

  rewrites() {
    const backendURL = (
      process.env.BACKEND_URL ?? "http://localhost:8080"
    ).replace(/\/+$/, "");

    const backofficeURL = (
      process.env.BACKOFFICE_URL ?? "http://localhost:8082"
    ).replace(/\/+$/, "");

    return [
      {
        source: "/backoffice/:path*",
        destination: `${backofficeURL}/backoffice/:path*`,
      },
      { source: "/api/v1/:path*", destination: `${backendURL}/v1/:path*` },
    ];
  },

  headers() {
    return [
      {
        source: "/:path*",
        headers: [
          {
            key: "X-Content-Type-Options",
            value: "nosniff",
          },
          {
            key: "Referrer-Policy",
            value: "strict-origin-when-cross-origin",
          },
          {
            key: "Content-Security-Policy",
            value: "frame-ancestors 'none'",
          },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(self), geolocation=()",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
