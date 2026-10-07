import { Heading, Text } from "react-email";
import Layout from "./_components/layout";

type InvitationAcceptedEmailProps = {
  memberName: string;
  organization: string;
  membershipRole: string;
};

export default function InvitationAcceptedEmail({
  memberName,
  organization,
  membershipRole,
}: InvitationAcceptedEmailProps) {
  return (
    <Layout preview="Your invitation was accepted">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        Your invitation was accepted
      </Heading>
      <Text>
        {memberName} accepted your invitation and joined {organization} as{" "}
        {membershipRole}.
      </Text>
      <Text>
        You can review their access in your organization’s member settings.
      </Text>
    </Layout>
  );
}

InvitationAcceptedEmail.PreviewProps = {
  memberName: "Alex",
  organization: "Acme",
  membershipRole: "member",
} satisfies InvitationAcceptedEmailProps;
