export function ProfileSettings() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Profile</h1>
      <p className="text-muted-foreground">
        Manage your name, email address, and account details.
      </p>
      <a href="/api/v1/auth/account" className="text-sm font-medium underline">
        Manage profile
      </a>
    </div>
  );
}
