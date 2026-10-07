import { Heading, Text } from "react-email";
import Layout from "./_components/layout";

type VoiceAgentFailureEmailProps = {
  agentName: string;
  organization: string;
  failureReason: string;
  occurredAt: string;
};

export default function VoiceAgentFailureEmail({
  agentName,
  organization,
  failureReason,
  occurredAt,
}: VoiceAgentFailureEmailProps) {
  return (
    <Layout preview="Voice agent needs attention">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        Voice agent needs attention
      </Heading>
      <Text>
        The voice agent {agentName} in {organization} encountered a terminal
        runtime failure at {occurredAt}.
      </Text>
      <Text>Reason: {failureReason}</Text>
      <Text>
        Review the affected session and provider configuration before retrying.
      </Text>
    </Layout>
  );
}

VoiceAgentFailureEmail.PreviewProps = {
  agentName: "Support agent",
  organization: "Acme",
  failureReason: "Speech provider unavailable",
  occurredAt: "12:30 UTC on 06 Oct 2026",
} satisfies VoiceAgentFailureEmailProps;
