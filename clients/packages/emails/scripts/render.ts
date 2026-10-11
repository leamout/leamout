import { mkdir, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { createElement, type ReactElement } from "react";
import { render, toPlainText } from "react-email";
import OrganizationInvitation, {
  subject as invitationSubject,
} from "../emails/organization-invitation";
import PasswordChanged, {
  subject as passwordSubject,
} from "../emails/password-changed";
import VerificationCode, {
  recoverySubject,
  subject as verificationSubject,
} from "../emails/verification-code";

const fields = {
  LEAMOUT_CODE_SENTINEL: "{{.Code}}",
  LEAMOUT_EXPIRY_SENTINEL: "{{.Expiry}}",
  LEAMOUT_INVITER_SENTINEL: "{{.Inviter}}",
  LEAMOUT_ORGANIZATION_SENTINEL: "{{.Organization}}",
  LEAMOUT_ROLE_SENTINEL: "{{.Role}}",
  "https://placeholder.invalid/invitation": "{{.AcceptURL}}",
  LEAMOUT_CHANGED_AT_SENTINEL: "{{.ChangedAt}}",
  "https://placeholder.invalid/recovery": "{{.RecoveryURL}}",
};

type Template = {
  name: string;
  subject: string;
  element: ReactElement;
  variants?: {
    condition: string;
    subject: string;
    element: ReactElement;
  }[];
};

const templates: Template[] = [
  {
    name: "verification-code",
    subject: verificationSubject,
    element: createElement(VerificationCode, {
      code: "LEAMOUT_CODE_SENTINEL",
      expiresAt: "LEAMOUT_EXPIRY_SENTINEL",
      purpose: "verification",
    }),
    variants: [
      {
        condition: 'eq .Purpose "password-recovery"',
        subject: recoverySubject,
        element: createElement(VerificationCode, {
          code: "LEAMOUT_CODE_SENTINEL",
          expiresAt: "LEAMOUT_EXPIRY_SENTINEL",
          purpose: "password-recovery",
        }),
      },
    ],
  },
  {
    name: "organization-invitation",
    subject: invitationSubject,
    element: createElement(OrganizationInvitation, {
      inviterName: "LEAMOUT_INVITER_SENTINEL",
      organizationName: "LEAMOUT_ORGANIZATION_SENTINEL",
      role: "LEAMOUT_ROLE_SENTINEL",
      acceptUrl: "https://placeholder.invalid/invitation",
      expiresAt: "LEAMOUT_EXPIRY_SENTINEL",
    }),
  },
  {
    name: "password-changed",
    subject: passwordSubject,
    element: createElement(PasswordChanged, {
      changedAt: "LEAMOUT_CHANGED_AT_SENTINEL",
      recoveryUrl: "https://placeholder.invalid/recovery",
    }),
  },
];

const output = fileURLToPath(
  new URL(
    "../../../../server/internal/platform/email/templates/",
    import.meta.url,
  ),
);
await mkdir(output, { recursive: true });

function substitute(content: string): string {
  for (const [sentinel, placeholder] of Object.entries(fields)) {
    content = content.replaceAll(sentinel, placeholder);
  }
  if (content.includes("SENTINEL") || content.includes("placeholder.invalid")) {
    throw new Error("Unresolved template placeholder");
  }
  return content;
}

for (const template of templates) {
  let subject = template.subject;
  let html = await render(template.element);
  let text = toPlainText(html);

  for (const variant of [...(template.variants ?? [])].reverse()) {
    const variantHTML = await render(variant.element);
    const branch = (matching: string, fallback: string) =>
      `{{if ${variant.condition}}}${matching}{{else}}${fallback}{{end}}`;
    subject = branch(variant.subject, subject);
    html = branch(variantHTML, html);
    text = branch(toPlainText(variantHTML), text);
  }

  await writeFile(`${output}${template.name}.subject.txt`, `${subject}\n`);
  await writeFile(`${output}${template.name}.html`, substitute(html));
  await writeFile(`${output}${template.name}.txt`, `${substitute(text)}\n`);
}
