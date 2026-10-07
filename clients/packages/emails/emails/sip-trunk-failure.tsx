import { Heading, Text } from "react-email";
import Layout from "./_components/layout";

type SIPTrunkFailureEmailProps = {
  trunkName: string;
  organization: string;
  failureReason: string;
  occurredAt: string;
};

export default function SIPTrunkFailureEmail({
  trunkName,
  organization,
  failureReason,
  occurredAt,
}: SIPTrunkFailureEmailProps) {
  return (
    <Layout preview="SIP trunk needs attention">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        SIP trunk needs attention
      </Heading>
      <Text>
        The SIP trunk {trunkName} in {organization} reported a failure at{" "}
        {occurredAt}.
      </Text>
      <Text>Reason: {failureReason}</Text>
      <Text>
        Review the trunk status, connectivity, and credentials. Calls using this
        trunk may be affected.
      </Text>
    </Layout>
  );
}

SIPTrunkFailureEmail.PreviewProps = {
  trunkName: "Primary trunk",
  organization: "Acme",
  failureReason: "SIP health check failed",
  occurredAt: "12:30 UTC on 06 Oct 2026",
} satisfies SIPTrunkFailureEmailProps;
