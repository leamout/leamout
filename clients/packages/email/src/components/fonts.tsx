import { Font } from "react-email";

const fontsourceBase = "https://cdn.jsdelivr.net/fontsource/fonts";

export function LeamoutFonts() {
  return (
    <>
      {[400, 500, 600, 700].map((weight) => (
        <Font
          key={`plex-sans-${weight}`}
          fontFamily="IBM Plex Sans"
          fallbackFontFamily={["Arial", "sans-serif"]}
          webFont={{
            url: `${fontsourceBase}/ibm-plex-sans@5.3.0/latin-${weight}-normal.woff2`,
            format: "woff2",
          }}
          fontWeight={weight}
          fontStyle="normal"
        />
      ))}
      {[400, 500, 600, 700].map((weight) => (
        <Font
          key={`plex-mono-${weight}`}
          fontFamily="IBM Plex Mono"
          fallbackFontFamily={["Courier New", "monospace"]}
          webFont={{
            url: `${fontsourceBase}/ibm-plex-mono@5.3.0/latin-${weight}-normal.woff2`,
            format: "woff2",
          }}
          fontWeight={weight}
          fontStyle="normal"
        />
      ))}
    </>
  );
}
