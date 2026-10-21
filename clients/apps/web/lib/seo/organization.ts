import { FOUNDER_ID } from "./person";
import { SITE_URL, type SchemaNode } from "./schema";

export function organizationSchema(): SchemaNode {
  return {
    "@type": "Organization",
    "@id": `${SITE_URL}/#organization`,
    name: "Leamout",
    url: SITE_URL,
    description:
      "Leamout is the carrier-grade runtime and control plane for autonomous voice agents.",
    founder: { "@id": FOUNDER_ID },
    sameAs: ["https://github.com/leamout"],
  };
}
