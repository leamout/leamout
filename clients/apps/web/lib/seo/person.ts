import type { Person } from "schema-dts";

export const FOUNDER_ID = "https://vidzro.com/#person";

export function founderSchema(): Person {
  return {
    "@type": "Person",
    "@id": FOUNDER_ID,
    name: "Coffey Vidzro",
    alternateName: "King Coffey Vidzro",
    url: "https://vidzro.com",
    jobTitle: "Founder",
    worksFor: { "@id": "https://leamout.com/#organization" },
    sameAs: [
      "https://github.com/coffeyvidzro",
      "https://www.linkedin.com/in/coffeyvidzro/",
      "https://x.com/coffeyvidzro",
    ],
  };
}
