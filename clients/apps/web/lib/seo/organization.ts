import type { Organization } from "schema-dts";

import { FOUNDER_ID } from "./person";
import { SITE_URL } from "./schema";

export function organizationSchema(): Organization {
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
