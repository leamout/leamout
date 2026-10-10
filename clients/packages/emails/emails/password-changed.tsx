import { Button, Heading, Text } from "react-email";
import Layout, {
  buttonStyle,
  headingStyle,
  textStyle,
} from "../components/layout";

type PasswordChangedProps = { changedAt: string; recoveryUrl: string };

export default function PasswordChanged({
  changedAt,
  recoveryUrl,
}: PasswordChangedProps) {
  return (
    <Layout preview="Your Leamout password was changed">
      <Heading style={headingStyle}>Your password was changed</Heading>
      <Text style={textStyle}>
        The password for your Leamout account was changed {changedAt}.
      </Text>
      <Text style={textStyle}>
        If you made this change, no further action is needed. If you did not,
        start password recovery to secure your account.
      </Text>
      <Button href={recoveryUrl} style={buttonStyle}>
        Start password recovery
      </Button>
    </Layout>
  );
}

PasswordChanged.PreviewProps = {
  changedAt: "October 10, 2026 at 07:30 UTC",
  recoveryUrl: "https://example.com/forgot-password",
} satisfies PasswordChangedProps;
