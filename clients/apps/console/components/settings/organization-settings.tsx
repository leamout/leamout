import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

export function OrganizationSettings() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Organization</h1>
      <p className="text-muted-foreground">
        Manage your organization name. Current details have not been loaded.
      </p>
      <PreviewSettingsForm
        fields={[
          {
            name: "organizationName",
            label: "Organization name",
            required: true,
          },
        ]}
      />
    </div>
  );
}
