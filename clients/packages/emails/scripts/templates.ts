import { createElement } from "react";
import OrganizationInvitation from "../emails/organization-invitation";
import VerificationCode from "../emails/verification-code";

// Register templates only once their Go delivery contract is available.
export const templates = [
  {
    name: "otp",
    element: createElement(VerificationCode, VerificationCode.ExportProps),
  },
  {
    name: "invitation",
    element: createElement(
      OrganizationInvitation,
      OrganizationInvitation.ExportProps,
    ),
  },
];
