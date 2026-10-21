import { serializeSchema, type SchemaNode } from "@/lib/seo/schema";

export function JsonLd({ nodes }: { nodes: SchemaNode[] }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: serializeSchema(nodes) }}
    />
  );
}
