import { Button, Heading, Text } from "react-email";
import Layout, {
  buttonStyle,
  headingStyle,
  textStyle,
} from "../components/layout";

type WelcomeProps = { name: string; consoleUrl: string };

export default function Welcome({ name, consoleUrl }: WelcomeProps) {
  return (
    <Layout preview="Your Leamout account is ready">
      <Heading style={headingStyle}>Welcome to Leamout</Heading>
      <Text style={textStyle}>Hi {name}, your account is ready.</Text>
      <Text style={textStyle}>
        Create your organization, configure an agent, and connect your existing
        phone numbers through a SIP trunk.
      </Text>
      <Button href={consoleUrl} style={buttonStyle}>
        Open console
      </Button>
    </Layout>
  );
}

Welcome.PreviewProps = {
  name: "Alex",
  consoleUrl: "https://example.com/console",
} satisfies WelcomeProps;
