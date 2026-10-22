import { serializeSchema } from "@/lib/seo/schema";
import type { Graph } from "schema-dts";

export function JsonLd({ nodes }: { nodes: Graph["@graph"] }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: serializeSchema(nodes) }}
    />
  );
}
