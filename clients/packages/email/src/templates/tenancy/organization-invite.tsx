import { Button, Section, Text } from "react-email";
import { EmailLayout } from "../../components/layout";

type OrganizationInviteEmailProps = {
  inviterName?: string;
  organizationName?: string;
  roleName?: string;
  acceptUrl?: string;
  expiresDays?: string;
};

export default function OrganizationInviteEmail({
  inviterName = "{{.InviterName}}",
  organizationName = "{{.OrganizationName}}",
  roleName = "{{.RoleName}}",
  acceptUrl = "{{.AcceptURL}}",
  expiresDays = "{{.ExpiresDays}}",
}: OrganizationInviteEmailProps) {
  return (
    <EmailLayout preview={`You're invited to ${organizationName}`}>
      <Section className="px-6 pt-8 pb-10 sm:px-10 sm:pt-10 sm:pb-12">
        <Text className="m-0 font-sans text-[48px] leading-none tracking-[-1.44px] text-fg">
          You&apos;re invited
        </Text>
        <Text className="m-0 mt-[18px] max-w-[470px] font-sans text-[14px] leading-[21px] text-fg-2">
          {inviterName} invited you to join {organizationName} on Leamout as{" "}
          {roleName}.
        </Text>
        <Button
          href={acceptUrl}
          className="mt-9 inline-block bg-brand px-5 py-3.5 text-center font-sans text-[15px] font-medium text-fg-inverted"
        >
          Accept invitation
        </Button>
        <Text className="m-0 mt-12 max-w-[350px] font-sans text-[11px] leading-[17px] text-fg-3">
          This invitation expires in {expiresDays} days. If you weren&apos;t
          expecting it, you can ignore this email.
        </Text>
      </Section>
    </EmailLayout>
  );
}

OrganizationInviteEmail.PreviewProps = {
  inviterName: "Alex",
  organizationName: "Acme Voice",
  roleName: "Admin",
  acceptUrl: "https://console.leamout.com/invitations/preview",
  expiresDays: "7",
} satisfies OrganizationInviteEmailProps;
