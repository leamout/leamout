import type { SoftwareApplication } from "schema-dts";

import { SITE_URL, siteUrl } from "./schema";

// Descriptive schema only. Do not invent ratings, downloads, prices, or a release status.
export function softwareSchema(): SoftwareApplication {
  return {
    "@type": "SoftwareApplication",
    "@id": `${SITE_URL}/#software`,
    name: "Leamout",
    url: siteUrl("/platform"),
    description:
      "SIP-native runtime and control plane for autonomous voice agents, with programmable call control, realtime media, and AI orchestration.",
    applicationCategory: "DeveloperApplication",
    operatingSystem: "Server",
    publisher: { "@id": `${SITE_URL}/#organization` },
  };
}
