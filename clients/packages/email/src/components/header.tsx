import { Text } from "react-email";

export function EmailHeader() {
  return <Text style={brand}>Leamout</Text>;
}

const brand = {
  margin: "0 0 32px",
  color: "#111111",
  fontSize: "18px",
  fontWeight: "700",
  letterSpacing: "-0.02em",
};
