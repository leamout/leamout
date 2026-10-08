import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  async rewrites() {
    const apiURL = process.env.API_URL ?? "http://localhost:8080";
    return [
      {
        source: "/api/v1/:path*",
        destination: `${apiURL.replace(/\/$/, "")}/v1/:path*`,
      },
    ];
  },
  transpilePackages: ["@leamout/ui"],
};

export default nextConfig;
