import type { TailwindConfig } from "react-email";

export const leamoutTailwindConfig: TailwindConfig = {
  theme: {
    extend: {
      colors: {
        canvas: "#FBFCFB",
        bg: "#FFFFFF",
        "bg-muted": "#F4F7F2",
        fg: "#103B05",
        "fg-2": "#315A27",
        "fg-3": "#7F957A",
        "fg-inverted": "#FBFFF9",
        stroke: "#D8E1D4",
        brand: "#103B05",
        danger: "#A13E32",
        "danger-muted": "#FFF3F0",
        warning: "#8A5D00",
        "warning-muted": "#FFF8E8",
      },
      boxShadow: {
        "matte-card":
          "0px 76px 21px 0px rgba(193,195,193,0), 0px 49px 19px 0px rgba(193,195,193,0.01), 0px 27px 16px 0px rgba(193,195,193,0.05), 0px 12px 12px 0px rgba(193,195,193,0.09), 0px 3px 7px 0px rgba(193,195,193,0.1)",
      },
      fontFamily: {
        sans: ["IBM Plex Sans", "Arial", "sans-serif"],
        mono: ["IBM Plex Mono", "Courier New", "monospace"],
      },
    },
  },
};
