import { Button, Heading, Link, Text } from "react-email";
import Layout from "./_components/layout";

type InvitationProps = {
  organization: string;
  inviter: string;
  membershipRole: string;
  acceptURL: string;
  expiresAt: string;
};

export default function InvitationEmail({
  organization,
  inviter,
  membershipRole,
  acceptURL,
  expiresAt,
}: InvitationProps) {
  return (
    <Layout preview="You’re invited to join an organization on Leamout">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        Join {organization}
      </Heading>
      <Text>
        {inviter} invited you to join {organization} on Leamout as{" "}
        {membershipRole}.
      </Text>
      <Button
        href={acceptURL}
        style={{
          backgroundColor: "#111827",
          borderRadius: "6px",
          color: "#ffffff",
          fontWeight: "bold",
          padding: "12px 20px",
          textDecoration: "none",
        }}
      >
        Accept invitation
      </Button>
      <Text>
        Or open this link: <Link href={acceptURL}>{acceptURL}</Link>
      </Text>
      <Text>Expires at {expiresAt}.</Text>
      <Text>
        If you were not expecting this invitation, you can ignore this email.
      </Text>
    </Layout>
  );
}

InvitationEmail.PreviewProps = {
  organization: "Acme",
  inviter: "Alice",
  membershipRole: "member",
  acceptURL: "https://leamout.com/invitations/example",
  expiresAt: "12:30 UTC on 06 Oct 2026",
} satisfies InvitationProps;
