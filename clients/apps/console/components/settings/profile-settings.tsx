import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

export function ProfileSettings() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Profile</h1>
      <p className="text-muted-foreground">
        Manage your display name. Account email and verification status have not
        been loaded.
      </p>
      <PreviewSettingsForm
        fields={[{ name: "displayName", label: "Display name" }]}
      />
    </div>
  );
}
