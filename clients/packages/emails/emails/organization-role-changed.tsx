import { Button, Heading, Text } from "react-email";
import Layout, {
  buttonStyle,
  headingStyle,
  textStyle,
} from "../components/layout";

type OrganizationRoleChangedProps = {
  organizationName: string;
  previousRole: string;
  newRole: string;
  organizationUrl: string;
};

export default function OrganizationRoleChanged({
  organizationName,
  previousRole,
  newRole,
  organizationUrl,
}: OrganizationRoleChangedProps) {
  return (
    <Layout preview={`Your role in ${organizationName} has changed`}>
      <Heading style={headingStyle}>Your organization role changed</Heading>
      <Text style={textStyle}>
        Your role in {organizationName} changed from {previousRole} to {newRole}
        .
      </Text>
      <Text style={textStyle}>
        Your available permissions may have changed. Contact an organization
        administrator if you have questions.
      </Text>
      <Button href={organizationUrl} style={buttonStyle}>
        Open organization
      </Button>
    </Layout>
  );
}

OrganizationRoleChanged.PreviewProps = {
  organizationName: "Acme",
  previousRole: "admin",
  newRole: "member",
  organizationUrl: "https://example.com/organization",
} satisfies OrganizationRoleChangedProps;
