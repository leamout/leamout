import type { BreadcrumbList, ListItem } from "schema-dts";

import { siteUrl } from "./schema";

export function breadcrumbsSchema(
  items: { name: string; path: string }[],
): BreadcrumbList {
  const itemListElement: ListItem[] = items.map((item, index) => ({
    "@type": "ListItem",
    position: index + 1,
    name: item.name,
    item: siteUrl(item.path),
  }));

  return {
    "@type": "BreadcrumbList",
    "@id": `${siteUrl(items[items.length - 1]?.path ?? "/")}#breadcrumb`,
    itemListElement,
  };
}
