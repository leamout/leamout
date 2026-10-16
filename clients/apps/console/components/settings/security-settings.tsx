export function SecuritySettings() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Security</h1>
      <p className="text-muted-foreground">
        Manage your password, authentication methods, and signed-in devices.
      </p>
      <a href="/api/v1/auth/account" className="text-sm font-medium underline">
        Manage account security
      </a>
    </div>
  );
}
