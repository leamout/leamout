import assert from "node:assert/strict";
import { existsSync, readFileSync, readdirSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = fileURLToPath(new URL("../", import.meta.url));
const marketing = path.join(root, "app", "(marketing)");
const read = (relativePath) =>
  readFileSync(path.join(root, relativePath), "utf8");

function visit(directory, onFile) {
  for (const entry of readdirSync(directory, { withFileTypes: true })) {
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      visit(fullPath, onFile);
    } else if (entry.isFile()) {
      onFile(fullPath);
    }
  }
}

const pages = new Map();
visit(marketing, (filePath) => {
  if (path.basename(filePath) !== "page.tsx") return;

  const relativePath = path.relative(marketing, path.dirname(filePath));
  if (relativePath.split(path.sep).some((part) => part.startsWith("["))) {
    return;
  }

  const route = relativePath ? `/${relativePath.split(path.sep).join("/")}` : "/";
  pages.set(route, filePath);
});

const sitemap = read("app/sitemap.ts");
const staticPathsMatch = sitemap.match(/const paths = \[([\s\S]*?)\];/);
assert.ok(staticPathsMatch, "Sitemap must declare its static paths");
const sitemapPaths = new Set(
  Array.from(staticPathsMatch[1].matchAll(/"([^"]+)"/g), (match) => match[1]),
);

const excludedPages = new Set([
  "/legal/privacy",
  "/legal/terms",
  "/legal/cookie",
  "/subprocessors",
]);

for (const [route, filePath] of pages) {
  if (excludedPages.has(route)) {
    assert.match(
      readFileSync(filePath, "utf8"),
      /noIndex:\s*true/,
      `${route} is pending publication and must remain noIndex`,
    );
  } else {
    assert.ok(sitemapPaths.has(route), `${route} is missing from sitemap`);
  }
}

for (const route of sitemapPaths) {
  assert.ok(pages.has(route), `Sitemap references missing page ${route}`);
  assert.ok(!excludedPages.has(route), `Unpublished page ${route} is in sitemap`);
}

assert.match(sitemap, /getPosts\(/, "Sitemap must include published blog posts");
assert.match(read("app/robots.ts"), /sitemap:/, "robots.txt must link the sitemap");

const nextConfig = read("next.config.ts");
for (const source of ["/docs/:path*", "/status/:path*"]) {
  assert.ok(
    nextConfig.includes(`source: "${source}"`),
    `Missing external redirect for ${source}`,
  );
}

const metadata = read("lib/metadata.ts");
const imageMatch = metadata.match(/\bimage\s*=\s*"([^"]+)"/);
assert.ok(imageMatch, "Default social image must be configured");
const imagePath = imageMatch[1];
const imageExists =
  imagePath === "/opengraph-image"
    ? existsSync(path.join(root, "app", "opengraph-image.tsx"))
    : existsSync(path.join(root, "public", imagePath.replace(/^\//, "")));
assert.ok(imageExists, `Missing default social image: ${imagePath}`);
for (const field of ["canonical:", "openGraph:", "twitter:"]) {
  assert.ok(metadata.includes(field), `Missing metadata field: ${field}`);
}

const missingLinks = [];
const staticHref = /\bhref\s*=\s*["'](\/[^"']*)["']/g;
for (const directory of ["app", "components"]) {
  visit(path.join(root, directory), (filePath) => {
    if (!/\.tsx?$/.test(filePath)) return;
    const source = readFileSync(filePath, "utf8");

    for (const match of source.matchAll(staticHref)) {
      const rawPath = match[1];
      if (rawPath.startsWith("//")) continue;
      const route = rawPath.split(/[?#]/)[0].replace(/\/$/, "") || "/";

      if (
        pages.has(route) ||
        route === "/docs" ||
        route.startsWith("/docs/") ||
        route === "/status" ||
        route.startsWith("/status/") ||
        existsSync(path.join(root, "public", route.slice(1)))
      ) {
        continue;
      }

      missingLinks.push(`${path.relative(root, filePath)}: ${route}`);
    }
  });
}
assert.deepEqual(missingLinks, [], "Broken static internal links detected");

console.log(
  `Website checks passed: ${pages.size} static pages, ${sitemapPaths.size} indexed routes, social image, and internal links.`,
);
