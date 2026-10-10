"use client";

import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

export function SecuritySettings() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Security</h1>
      <p className="text-muted-foreground">
        Password change preview. Session management and additional security
        methods are deferred.
      </p>
      <PreviewSettingsForm
        action="Change password"
        fields={[
          {
            name: "currentPassword",
            label: "Current password",
            type: "password",
            required: true,
          },
          {
            name: "newPassword",
            label: "New password",
            type: "password",
            required: true,
          },
          {
            name: "confirmPassword",
            label: "Confirm new password",
            type: "password",
            required: true,
          },
        ]}
        validate={(values) =>
          values.newPassword === values.confirmPassword
            ? []
            : [{ field: "confirmPassword", message: "Passwords must match." }]
        }
      />
    </div>
  );
}
