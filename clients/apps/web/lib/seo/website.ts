import type { WebSite } from "schema-dts";

import { SITE_URL } from "./schema";

export function websiteSchema(): WebSite {
  return {
    "@type": "WebSite",
    "@id": `${SITE_URL}/#website`,
    name: "Leamout",
    url: SITE_URL,
    inLanguage: "en",
    publisher: { "@id": `${SITE_URL}/#organization` },
  };
}
