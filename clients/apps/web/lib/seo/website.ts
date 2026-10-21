import { SITE_URL, type SchemaNode } from "./schema";

export function websiteSchema(): SchemaNode {
  return {
    "@type": "WebSite",
    "@id": `${SITE_URL}/#website`,
    name: "Leamout",
    url: SITE_URL,
    inLanguage: "en",
    publisher: { "@id": `${SITE_URL}/#organization` },
  };
}
