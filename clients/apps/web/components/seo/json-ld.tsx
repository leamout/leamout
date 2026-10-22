import type { Graph } from "schema-dts";
import { serializeSchema } from "@/lib/seo/schema";

export function JsonLd({ nodes }: { nodes: Graph["@graph"] }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: serializeSchema(nodes) }}
    />
  );
}
