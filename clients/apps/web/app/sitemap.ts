import { existsSync, readdirSync } from "node:fs";
import { join, relative, sep } from "node:path";
import type { MetadataRoute } from "next";

const BASE_URL = "https://leamout.com";
const MARKETING_DIR = join(process.cwd(), "app", "(marketing)");

function getRoutes(directory: string): string[] {
  if (!existsSync(directory)) {
    return [];
  }

  const routes: string[] = [];

  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    if (!entry.isDirectory()) {
      continue;
    }

    if (
      entry.name.startsWith("_") ||
      entry.name.startsWith("@") ||
      entry.name.startsWith("[")
    ) {
      continue;
    }

    const directoryPath = join(directory, entry.name);
    const pagePath = join(directoryPath, "page.tsx");

    if (existsSync(pagePath)) {
      const pathname = relative(MARKETING_DIR, directoryPath)
        .split(sep)
        .filter((segment) => !segment.startsWith("("))
        .join("/");

      routes.push(`/${pathname}`);
    }

    routes.push(...getRoutes(directoryPath));
  }

  return routes;
}

export default function sitemap(): MetadataRoute.Sitemap {
  const routes = ["/", ...getRoutes(MARKETING_DIR)];

  return [...new Set(routes)].sort().map((route) => ({
    url: new URL(route, BASE_URL).toString(),
  }));
}
