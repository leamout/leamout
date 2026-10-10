import { mkdir, writeFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { createElement, type ReactElement } from "react";
import { render } from "react-email";
import OrganizationInvitation from "../emails/organization-invitation";
import OrganizationRoleChanged from "../emails/organization-role-changed";
import PasswordChanged from "../emails/password-changed";
import VerificationCode from "../emails/verification-code";
import Welcome from "../emails/welcome";

function object(value: unknown): Record<string, unknown> {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error("Expected a JSON object.");
  }
  return value as Record<string, unknown>;
}

function text(props: Record<string, unknown>, key: string): string {
  const value = props[key];
  if (typeof value !== "string" || !value.trim()) {
    throw new Error(`${key} must be a nonempty string.`);
  }
  return value;
}

function url(props: Record<string, unknown>, key: string): string {
  const value = text(props, key);
  let parsed: URL;
  try {
    parsed = new URL(value);
  } catch {
    throw new Error(`${key} must be an absolute HTTP or HTTPS URL.`);
  }
  if (
    !["http:", "https:"].includes(parsed.protocol) ||
    parsed.username ||
    parsed.password
  ) {
    throw new Error(`${key} must be an HTTP or HTTPS URL without credentials.`);
  }
  return value;
}

export async function renderEmail(
  input: unknown,
): Promise<{ html: string; text: string }> {
  const request = object(input);
  const template = text(request, "template");
  const props = object(request.props);
  let element: ReactElement;

  switch (template) {
    case "welcome":
      element = createElement(Welcome, {
        name: text(props, "name"),
        consoleUrl: url(props, "consoleUrl"),
      });
      break;
    case "verification-code": {
      const purpose = text(props, "purpose");
      if (
        purpose !== "create-account" &&
        purpose !== "log-in" &&
        purpose !== "password-recovery"
      ) {
        throw new Error(
          "purpose must be create-account, log-in, or password-recovery.",
        );
      }
      const expiresInMinutes = props.expiresInMinutes;
      if (
        typeof expiresInMinutes !== "number" ||
        !Number.isSafeInteger(expiresInMinutes) ||
        expiresInMinutes < 1
      ) {
        throw new Error("expiresInMinutes must be a positive whole number.");
      }
      element = createElement(VerificationCode, {
        code: text(props, "code"),
        purpose,
        expiresInMinutes,
      });
      break;
    }
    case "organization-invitation":
      element = createElement(OrganizationInvitation, {
        inviterName: text(props, "inviterName"),
        organizationName: text(props, "organizationName"),
        role: text(props, "role"),
        acceptUrl: url(props, "acceptUrl"),
        expiresAt: text(props, "expiresAt"),
      });
      break;
    case "password-changed":
      element = createElement(PasswordChanged, {
        changedAt: text(props, "changedAt"),
        recoveryUrl: url(props, "recoveryUrl"),
      });
      break;
    case "organization-role-changed":
      element = createElement(OrganizationRoleChanged, {
        organizationName: text(props, "organizationName"),
        previousRole: text(props, "previousRole"),
        newRole: text(props, "newRole"),
        organizationUrl: url(props, "organizationUrl"),
      });
      break;
    default:
      throw new Error("Unsupported email template.");
  }

  const html = await render(element);
  const plainText = await render(element, { plainText: true });
  return { html, text: plainText };
}

// Export only templates with existing Go delivery contracts. Sentinels are
// replaced after React rendering so HTML and text retain Go template actions.
export async function exportTemplates() {
  const output = fileURLToPath(
    new URL(
      "../../../../server/internal/platform/email/templates/",
      import.meta.url,
    ),
  );
  const expiresAt = "LEAMOUT_EXPIRY_SENTINEL";
  const definitions = [
    {
      name: "otp",
      element: createElement(VerificationCode, {
        code: "LEAMOUT_CODE_SENTINEL",
        expiresAt,
        purpose: "log-in",
      }),
      fields: {
        LEAMOUT_CODE_SENTINEL: "{{.Code}}",
        LEAMOUT_EXPIRY_SENTINEL:
          '{{.ExpiresAt.UTC.Format "15:04 UTC on 02 Jan 2006"}}',
      },
    },
    {
      name: "invitation",
      element: createElement(OrganizationInvitation, {
        inviterName: "LEAMOUT_INVITER_SENTINEL",
        organizationName: "LEAMOUT_ORGANIZATION_SENTINEL",
        role: "LEAMOUT_ROLE_SENTINEL",
        acceptUrl: "https://leamout.invalid/accept-sentinel",
        expiresAt,
      }),
      fields: {
        LEAMOUT_INVITER_SENTINEL: "{{.Inviter}}",
        LEAMOUT_ORGANIZATION_SENTINEL: "{{.Organization}}",
        LEAMOUT_ROLE_SENTINEL: "{{.Role}}",
        "https://leamout.invalid/accept-sentinel": "{{.AcceptURL}}",
        LEAMOUT_EXPIRY_SENTINEL:
          '{{.ExpiresAt.UTC.Format "15:04 UTC on 02 Jan 2006"}}',
      },
    },
  ];
  await mkdir(output, { recursive: true });
  for (const definition of definitions) {
    for (const [extension, plainText] of [
      ["html", false],
      ["txt", true],
    ] as const) {
      let body = plainText
        ? await render(definition.element, { plainText: true })
        : await render(definition.element);
      for (const [sentinel, action] of Object.entries(definition.fields)) {
        if (!body.includes(sentinel))
          throw new Error(
            `Missing export field in ${definition.name}.${extension}.`,
          );
        body = body.replaceAll(sentinel, action);
      }
      await writeFile(
        `${output}${definition.name}.${extension}`,
        `{{/* Generated by clients/packages/emails/scripts/render.ts. Do not edit. */}}\n${body}\n`,
      );
    }
  }
}

if (import.meta.main) {
  try {
    if (process.argv[2] === "--export") {
      await exportTemplates();
    } else {
      const source = await Bun.stdin.text();
      let input: unknown;
      try {
        input = JSON.parse(source);
      } catch {
        throw new Error("Input must be valid JSON.");
      }
      const result = await renderEmail(input);
      process.stdout.write(`${JSON.stringify(result)}\n`);
    }
  } catch (error) {
    console.error(
      error instanceof Error ? error.message : "Email rendering failed.",
    );
    process.exitCode = 1;
  }
}
