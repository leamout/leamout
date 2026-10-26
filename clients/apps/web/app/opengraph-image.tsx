import { ImageResponse } from "next/og";

export const alt = "Leamout — Voice agents. Built for real calls.";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpenGraphImage() {
  return new ImageResponse(
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        width: "100%",
        height: "100%",
        padding: 72,
        backgroundColor: "#101316",
        color: "#F5F6F4",
        fontFamily: "Arial, sans-serif",
      }}
    >
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            fontSize: 36,
            fontWeight: 700,
            letterSpacing: -1,
          }}
        >
          <span>Leamout</span>
          <span style={{ color: "#F43E01" }}>.</span>
        </div>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <span
            style={{
              width: 10,
              height: 10,
              borderRadius: 10,
              backgroundColor: "#F43E01",
            }}
          />
          <span
            style={{
              fontSize: 17,
              letterSpacing: 2.2,
              color: "#C3C8C7",
            }}
          >
            SIP-NATIVE INFRASTRUCTURE
          </span>
        </div>
      </div>

      <div
        style={{
          display: "flex",
          flex: 1,
          flexDirection: "column",
          justifyContent: "center",
          gap: 25,
        }}
      >
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            fontSize: 76,
            fontWeight: 700,
            letterSpacing: -3,
            lineHeight: 1.05,
          }}
        >
          <span>Voice agents.</span>
          <span style={{ color: "#F43E01" }}>Built for real calls.</span>
        </div>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 4,
            fontSize: 27,
            lineHeight: 1.3,
            color: "#C3C8C7",
          }}
        >
          <span>Carrier-grade runtime and control plane</span>
          <span>for autonomous voice agents.</span>
        </div>
      </div>

      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          paddingTop: 24,
          borderTop: "1px solid #424A4E",
          fontSize: 18,
        }}
      >
        <span style={{ color: "#C3C8C7" }}>One managed runtime. Your providers.</span>
        <span style={{ color: "#F5F6F4" }}>leamout.com</span>
      </div>
    </div>,
    size,
  );
}
