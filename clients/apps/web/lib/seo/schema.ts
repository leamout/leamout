export const SITE_URL = "https://leamout.com";

export type SchemaNode = Record<string, unknown> & {
  "@type": string;
  "@id"?: string;
};

export function siteUrl(path = "/"): string {
  return new URL(path, SITE_URL).toString();
}

export function schemaGraph(nodes: SchemaNode[]) {
  return {
    "@context": "https://schema.org",
    "@graph": nodes,
  };
}

// Escape script-ending sequences before embedding the JSON in HTML.
export function serializeSchema(nodes: SchemaNode[]): string {
  return JSON.stringify(schemaGraph(nodes)).replace(/</g, "\\u003c");
}
