import { mkdir, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { createElement } from "react";
import { render, toPlainText } from "react-email";
import OrganizationInvitation from "../emails/organization-invitation";
import PasswordChanged from "../emails/password-changed";
import VerificationCode from "../emails/verification-code";

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

const templates = [
  {
    name: "verification-code",
    element: createElement(VerificationCode, {
      code: "LEAMOUT_CODE_SENTINEL",
      expiresAt: "LEAMOUT_EXPIRY_SENTINEL",
      purpose: "verification",
    }),
  },
  {
    name: "organization-invitation",
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
  const html = await render(template.element);
  const text = toPlainText(html);
  await writeFile(`${output}${template.name}.html`, substitute(html));
  await writeFile(`${output}${template.name}.txt`, `${substitute(text)}\n`);
}
