import type { AboutPage, WebPage } from "schema-dts";

import { SITE_URL, siteUrl } from "./schema";

type PageOptions = {
  path: string;
  name: string;
  description: string;
  type?: "WebPage" | "AboutPage";
};

export function webpageSchema({
  path,
  name,
  description,
  type = "WebPage",
}: PageOptions): WebPage | AboutPage {
  const url = siteUrl(path);
  const fields = {
    "@id": `${url}#webpage`,
    url,
    name,
    description,
    inLanguage: "en",
    isPartOf: { "@id": `${SITE_URL}/#website` },
    about: { "@id": `${SITE_URL}/#organization` },
  };

  if (type === "AboutPage") {
    return { "@type": "AboutPage", ...fields };
  }

  return { "@type": "WebPage", ...fields };
}
