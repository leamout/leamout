import type { Graph } from "schema-dts";

export const SITE_URL = "https://leamout.com";

export function siteUrl(path = "/"): string {
  return new URL(path, SITE_URL).toString();
}

export function schemaGraph(nodes: Graph["@graph"]): Graph {
  return {
    "@context": "https://schema.org",
    "@graph": nodes,
  };
}

// Escape angle brackets before embedding JSON-LD in an HTML script element.
export function serializeSchema(nodes: Graph["@graph"]): string {
  return JSON.stringify(schemaGraph(nodes)).replace(/</g, "\\u003c");
}
