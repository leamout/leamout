import { siteUrl, type SchemaNode } from "./schema";

export function breadcrumbsSchema(
  items: { name: string; path: string }[],
): SchemaNode {
  return {
    "@type": "BreadcrumbList",
    "@id": `${siteUrl(items[items.length - 1]?.path ?? "/")}#breadcrumb`,
    itemListElement: items.map((item, index) => ({
      "@type": "ListItem",
      position: index + 1,
      name: item.name,
      item: siteUrl(item.path),
    })),
  };
}
