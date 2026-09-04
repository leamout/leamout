import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type RoleChangedEmailProps = {
  organizationName?: string;
  roleName?: string;
  changedByName?: string;
  organizationUrl?: string;
};

export default function RoleChangedEmail({
  organizationName = "{{.OrganizationName}}",
  roleName = "{{.RoleName}}",
  changedByName = "{{.ChangedByName}}",
  organizationUrl = "{{.OrganizationURL}}",
}: RoleChangedEmailProps) {
  return (
    <EmailLayout preview={`Your role in ${organizationName} changed`}>
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[40px] leading-[1.05] tracking-[-1.2px] text-fg">
          Your role changed
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          {changedByName} changed your role in {organizationName} to {roleName}.
        </Text>
        <Button
          href={organizationUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Open organization
        </Button>
      </Section>
    </EmailLayout>
  );
}

RoleChangedEmail.PreviewProps = {
  organizationName: "Acme Voice",
  roleName: "Developer",
  changedByName: "Jordan",
  organizationUrl: "https://console.leamout.com/org/acme",
} satisfies RoleChangedEmailProps;
