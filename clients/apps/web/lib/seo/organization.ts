import { SITE_URL, type SchemaNode } from "./schema";

export function organizationSchema(): SchemaNode {
  return {
    "@type": "Organization",
    "@id": `${SITE_URL}/#organization`,
    name: "Leamout",
    url: SITE_URL,
    description:
      "Leamout is the carrier-grade runtime and control plane for autonomous voice agents.",
    sameAs: ["https://github.com/leamout"],
  };
}
