import { Heading, Text } from "react-email";
import Layout from "./_components/layout";

type SecurityAlertEmailProps = {
  eventName: string;
  occurredAt: string;
};

export default function SecurityAlertEmail({
  eventName,
  occurredAt,
}: SecurityAlertEmailProps) {
  return (
    <Layout preview="Account security alert">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        Account security alert
      </Heading>
      <Text>
        {eventName} occurred on your Leamout account at {occurredAt}.
      </Text>
      <Text>
        If you made this change, no action is needed. If you do not recognize
        it, secure your account and contact your organization administrator.
      </Text>
    </Layout>
  );
}

SecurityAlertEmail.PreviewProps = {
  eventName: "Password changed",
  occurredAt: "12:30 UTC on 06 Oct 2026",
} satisfies SecurityAlertEmailProps;
