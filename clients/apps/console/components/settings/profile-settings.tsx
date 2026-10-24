export function ProfileSettings() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Profile</h1>
      <p className="text-muted-foreground">
        Update your identity and personal details securely through Kinde.
        Leamout organization settings are managed separately.
      </p>
      <a
        href="/api/v1/auth/account"
        className="inline-flex h-9 items-center rounded-md border px-4 text-sm font-medium hover:bg-accent"
      >
        Manage personal profile
      </a>
    </div>
  );
}
