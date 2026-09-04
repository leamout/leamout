import type { CSSProperties } from "react";
import { Hr, Text } from "react-email";

export function EmailFooter() {
  return (
    <>
      <Hr style={divider} />
      <Text style={footer}>Leamout · Self-hosted programmable communications infrastructure</Text>
    </>
  );
}

const divider: CSSProperties = {
  borderColor: "#e5e5e5",
  margin: "32px 0 20px",
};

const footer: CSSProperties = {
  margin: 0,
  color: "#737373",
  fontSize: "12px",
  lineHeight: "18px",
};
