import {
  Body,
  Container,
  Head,
  Heading,
  Html,
  Preview,
  Text,
} from "react-email";

export default function Welcome({ name = "Alex" }: { name?: string }) {
  return (
    <Html>
      <Head />
      <Preview>Welcome to Leamout</Preview>
      <Body
        style={{
          backgroundColor: "#f8fafc",
          fontFamily: "Arial, sans-serif",
          padding: "32px 16px",
        }}
      >
        <Container
          style={{
            backgroundColor: "#ffffff",
            padding: "32px",
            borderRadius: "12px",
            maxWidth: "560px",
          }}
        >
          <Heading>Welcome to Leamout, {name}</Heading>
          <Text>Your workspace for autonomous voice agents.</Text>
          <Text>
            Connect your AI providers and SIP trunk, then configure your first
            agent.
          </Text>
        </Container>
      </Body>
    </Html>
  );
}
