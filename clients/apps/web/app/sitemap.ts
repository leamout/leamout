import type { MetadataRoute } from "next";

const paths = [
  "/",
  "/platform",
  "/integrations",
  "/pricing",
  "/about",
  "/security",
  "/contact",
  "/brand",
];

export default function sitemap(): MetadataRoute.Sitemap {
  return paths.map((path) => ({
    url: new URL(path, "https://leamout.com").toString(),
  }));
}
