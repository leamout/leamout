import { SettingsNavigation } from "@/components/settings/settings-navigation";

export default function SettingsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-8">
      <SettingsNavigation />
      {children}
    </div>
  );
}
