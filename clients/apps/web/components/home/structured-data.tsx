const SITE_URL = "https://leamout.com";

const graph = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "Organization",
      "@id": `${SITE_URL}/#organization`,
      name: "Leamout",
      url: SITE_URL,
      description:
        "Leamout is the carrier-grade runtime and control plane for autonomous voice agents.",
    },
    {
      "@type": "WebSite",
      "@id": `${SITE_URL}/#website`,
      url: SITE_URL,
      name: "Leamout",
      description:
        "SIP-native infrastructure for autonomous voice agents.",
      publisher: { "@id": `${SITE_URL}/#organization` },
      inLanguage: "en",
    },
    {
      "@type": "WebPage",
      "@id": `${SITE_URL}/#webpage`,
      url: `${SITE_URL}/`,
      name: "Leamout — Runtime and Control Plane for Voice Agents",
      description:
        "SIP-native infrastructure for autonomous voice agents, including programmable call control, real-time media, and AI orchestration.",
      isPartOf: { "@id": `${SITE_URL}/#website` },
      about: { "@id": `${SITE_URL}/#organization` },
      inLanguage: "en",
    },
  ],
};

export function HomeStructuredData() {
  // Escape angle brackets so schema content cannot terminate its script element.
  const json = JSON.stringify(graph).replace(/</g, "\\u003c");

  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: json }}
    />
  );
}
