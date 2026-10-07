import { Heading, Text } from "react-email";
import Layout from "./_components/layout";

type WelcomeEmailProps = {
  userName: string;
};

export default function WelcomeEmail({ userName }: WelcomeEmailProps) {
  return (
    <Layout preview="Welcome to Leamout">
      <Heading as="h1" style={{ fontSize: "22px" }}>
        Welcome to Leamout
      </Heading>
      <Text>Hello {userName},</Text>
      <Text>
        Welcome to Leamout. You can now create an organization, configure your
        SIP trunks, and build voice agents.
      </Text>
      <Text>Your account is ready. Sign in to get started.</Text>
    </Layout>
  );
}

WelcomeEmail.PreviewProps = {
  userName: "Alex",
} satisfies WelcomeEmailProps;
