import { Button, Heading, Text } from "react-email";
import Layout, {
  buttonStyle,
  headingStyle,
  textStyle,
} from "../components/layout";

type OrganizationInvitationProps = {
  inviterName: string;
  organizationName: string;
  role: string;
  acceptUrl: string;
  expiresAt: string;
};

export default function OrganizationInvitation({
  inviterName,
  organizationName,
  role,
  acceptUrl,
  expiresAt,
}: OrganizationInvitationProps) {
  return (
    <Layout preview={`You have been invited to ${organizationName}`}>
      <Heading style={headingStyle}>Join {organizationName}</Heading>
      <Text style={textStyle}>
        {inviterName} invited you to join {organizationName} on Leamout as{" "}
        {role}.
      </Text>
      <Button href={acceptUrl} style={buttonStyle}>
        Accept invitation
      </Button>
      <Text style={textStyle}>This invitation expires {expiresAt}.</Text>
      <Text style={textStyle}>
        If you were not expecting this invitation, you can ignore it.
      </Text>
    </Layout>
  );
}

OrganizationInvitation.PreviewProps = {
  inviterName: "Alex",
  organizationName: "Acme",
  role: "member",
  acceptUrl: "https://example.com/invitation",
  expiresAt: "October 17, 2026 at 12:00 UTC",
} satisfies OrganizationInvitationProps;
