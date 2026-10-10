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

if (import.meta.main) {
  try {
    const source = await Bun.stdin.text();
    let input: unknown;
    try {
      input = JSON.parse(source);
    } catch {
      throw new Error("Input must be valid JSON.");
    }
    const result = await renderEmail(input);
    process.stdout.write(`${JSON.stringify(result)}\n`);
  } catch (error) {
    console.error(
      error instanceof Error ? error.message : "Email rendering failed.",
    );
    process.exitCode = 1;
  }
}
