import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

type Member = { id: string; name: string; role: string; status: string };

export function MembersSettings({ members = [] }: { members?: Member[] }) {
  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">Members</h1>
        <p className="text-muted-foreground">
          Manage organization access. Member roles include owner, admin, and
          member.
        </p>
      </div>
      {members.length ? (
        <ul className="space-y-3">
          {members.map((member) => (
            <li key={member.id} className="rounded-lg border p-4">
              <p className="font-medium">{member.name}</p>
              <p className="text-sm text-muted-foreground">
                {member.role} · {member.status}
              </p>
            </li>
          ))}
        </ul>
      ) : (
        <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
          Members and pending invitations have not been loaded.
        </p>
      )}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Invite member</h2>
        <p className="text-sm text-muted-foreground">
          Invitation and role assignment are preview-only. No email will be
          sent.
        </p>
        <PreviewSettingsForm
          action="Invite member"
          fields={[
            {
              name: "inviteRole",
              label: "Role",
              required: true,
              options: ["admin", "member"],
            },
            {
              name: "inviteEmail",
              label: "Email address",
              type: "email",
              required: true,
            },
          ]}
        />
      </section>
    </div>
  );
}
