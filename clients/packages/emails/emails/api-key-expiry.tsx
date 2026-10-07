import { Heading, Text } from "react-email";
import Layout from "./_components/layout";

type APIKeyExpiryEmailProps = {
  keyName: string;
  organization: string;
  keyExpiresAt: string;
};

export default function APIKeyExpiryEmail({
  keyName,
  organization,
  keyExpiresAt,
}: APIKeyExpiryEmailProps) {
  return (
    <Layout preview="Your API key is expiring">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        Your API key is expiring
      </Heading>
      <Text>
        The API key {keyName} for {organization} expires at {keyExpiresAt}.
      </Text>
      <Text>
        Create a replacement key, update the applications using it, and revoke
        the old key when the transition is complete.
      </Text>
      <Text>This email never contains your API key secret.</Text>
    </Layout>
  );
}

APIKeyExpiryEmail.PreviewProps = {
  keyName: "Production integration",
  organization: "Acme",
  keyExpiresAt: "12:30 UTC on 13 Oct 2026",
} satisfies APIKeyExpiryEmailProps;
