"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@leamout/ui/components/button";
import { Field, FieldError, FieldLabel } from "@leamout/ui/components/field";
import { Input } from "@leamout/ui/components/input";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";
import { z } from "zod";
import { useOrganization } from "@/components/onboarding/organization-provider";
import { api, errorMessage } from "@/lib/api";

type Member = {
  user_id: string;
  role: "owner" | "admin" | "member";
  status: string;
  created_at: string;
};

type Invitation = {
  id: string;
  email: string;
  role: string;
  status: string;
  expires_at: string;
};

const schema = z.object({
  email: z.email("Enter a valid email address."),
  role: z.enum(["admin", "member"]),
});

export function MembersSettings() {
  const { selected } = useOrganization();
  const organizationId = selected?.id ?? "";
  const canManage = selected?.role === "owner" || selected?.role === "admin";
  const client = useQueryClient();
  const [removing, setRemoving] = useState<string | null>(null);
  const [revoking, setRevoking] = useState<string | null>(null);
  const [editing, setEditing] = useState<string | null>(null);
  const form = useForm<z.infer<typeof schema>>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", role: "member" },
  });
  const base = `/organizations/${encodeURIComponent(organizationId)}`;
  const members = useQuery({
    queryKey: ["organization-members", organizationId],
    queryFn: ({ signal }) =>
      api<Member[]>(`${base}/members/`, { organizationId, signal }),
    enabled: Boolean(organizationId),
  });
  const invitations = useQuery({
    queryKey: ["organization-invitations", organizationId],
    queryFn: ({ signal }) =>
      api<Invitation[]>(`${base}/invitations/`, { organizationId, signal }),
    enabled: Boolean(organizationId) && canManage,
  });
  const invite = useMutation({
    mutationFn: (values: z.infer<typeof schema>) =>
      api<Invitation>(`${base}/invitations/`, {
        method: "POST",
        organizationId,
        body: values,
      }),
    onSuccess: async () => {
      form.reset();
      await client.invalidateQueries({
        queryKey: ["organization-invitations", organizationId],
      });
    },
  });
  const update = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      api<Member>(`${base}/members/${encodeURIComponent(userId)}`, {
        method: "PATCH",
        organizationId,
        body: { role },
      }),
    onSuccess: async () => {
      setEditing(null);
      await client.invalidateQueries({
        queryKey: ["organization-members", organizationId],
      });
    },
  });
  const remove = useMutation({
    mutationFn: (userId: string) =>
      api<void>(`${base}/members/${encodeURIComponent(userId)}`, {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      setRemoving(null);
      await client.invalidateQueries({
        queryKey: ["organization-members", organizationId],
      });
    },
  });
  const revoke = useMutation({
    mutationFn: (id: string) =>
      api<void>(`${base}/invitations/${encodeURIComponent(id)}`, {
        method: "DELETE",
        organizationId,
      }),
    onSuccess: async () => {
      setRevoking(null);
      await client.invalidateQueries({
        queryKey: ["organization-invitations", organizationId],
      });
    },
  });

  return (
    <div className="space-y-8">
      <header className="space-y-2">
        <h1 className="text-2xl font-semibold">Members</h1>
        <p className="text-muted-foreground">
          Manage access and email invitations for your organization.
        </p>
      </header>

      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Members</h2>
        {members.isPending ? (
          <p role="status">Loading members…</p>
        ) : members.isError ? (
          <div role="alert" className="space-y-3">
            <p>{errorMessage(members.error)}</p>
            <Button variant="outline" onClick={() => void members.refetch()}>
              Retry
            </Button>
          </div>
        ) : members.data.length ? (
          <ul className="space-y-3">
            {members.data.map((member) => (
              <li
                key={member.user_id}
                className="space-y-3 rounded-lg border p-4"
              >
                <p className="break-all font-mono text-xs">{member.user_id}</p>
                <p className="text-sm text-muted-foreground">
                  {member.role} · {member.status}
                </p>
                {canManage && member.role !== "owner" && (
                  <div className="flex flex-wrap gap-2">
                    {editing === member.user_id ? (
                      <>
                        <Button
                          variant="outline"
                          disabled={update.isPending}
                          onClick={() =>
                            update.mutate({
                              userId: member.user_id,
                              role:
                                member.role === "admin" ? "member" : "admin",
                            })
                          }
                        >
                          Set {member.role === "admin" ? "member" : "admin"}
                        </Button>
                        <Button
                          variant="outline"
                          onClick={() => setEditing(null)}
                        >
                          Cancel
                        </Button>
                      </>
                    ) : (
                      <Button
                        variant="outline"
                        onClick={() => setEditing(member.user_id)}
                      >
                        Change role…
                      </Button>
                    )}
                    {removing === member.user_id ? (
                      <>
                        <Button
                          variant="destructive"
                          disabled={remove.isPending}
                          onClick={() => remove.mutate(member.user_id)}
                        >
                          Confirm removal
                        </Button>
                        <Button
                          variant="outline"
                          onClick={() => setRemoving(null)}
                        >
                          Cancel
                        </Button>
                      </>
                    ) : (
                      <Button
                        variant="outline"
                        onClick={() => setRemoving(member.user_id)}
                      >
                        Remove…
                      </Button>
                    )}
                  </div>
                )}
              </li>
            ))}
          </ul>
        ) : (
          <p>No members found.</p>
        )}
        {(remove.isError || update.isError) && (
          <p role="alert" className="text-sm text-destructive">
            {errorMessage(remove.error ?? update.error)}
          </p>
        )}
      </section>

      {canManage && (
        <>
          <section className="space-y-4">
            <h2 className="text-lg font-semibold">Invite member</h2>
            <form
              className="max-w-lg space-y-4"
              noValidate
              onSubmit={form.handleSubmit((values) => invite.mutate(values))}
            >
              <Controller
                name="email"
                control={form.control}
                render={({ field, fieldState }) => (
                  <Field data-invalid={fieldState.invalid}>
                    <FieldLabel htmlFor="invite-email">
                      Email address
                    </FieldLabel>
                    <Input
                      {...field}
                      id="invite-email"
                      type="email"
                      aria-invalid={fieldState.invalid}
                      required
                    />
                    {fieldState.invalid && (
                      <FieldError errors={[fieldState.error]} />
                    )}
                  </Field>
                )}
              />
              <Field>
                <FieldLabel htmlFor="invite-role">Role</FieldLabel>
                <select
                  id="invite-role"
                  className="h-10 rounded-md border bg-background px-3"
                  {...form.register("role")}
                >
                  <option value="member">Member</option>
                  <option value="admin">Admin</option>
                </select>
              </Field>
              <Button type="submit" disabled={invite.isPending}>
                {invite.isPending ? "Sending…" : "Send invitation"}
              </Button>
              {invite.isSuccess && (
                <p role="status">
                  Invitation created and queued for email delivery.
                </p>
              )}
              {invite.isError && (
                <p role="alert" className="text-sm text-destructive">
                  {errorMessage(invite.error)}
                </p>
              )}
            </form>
          </section>

          <section className="space-y-4">
            <h2 className="text-lg font-semibold">Pending invitations</h2>
            {invitations.isPending ? (
              <p role="status">Loading invitations…</p>
            ) : invitations.isError ? (
              <p role="alert">{errorMessage(invitations.error)}</p>
            ) : invitations.data.length ? (
              <ul className="space-y-3">
                {invitations.data.map((invitation) => (
                  <li
                    key={invitation.id}
                    className="flex flex-wrap items-center justify-between gap-3 rounded-lg border p-4"
                  >
                    <div>
                      <p className="font-medium">{invitation.email}</p>
                      <p className="text-sm text-muted-foreground">
                        {invitation.role} · Expires{" "}
                        {new Date(invitation.expires_at).toLocaleDateString()}
                      </p>
                    </div>
                    {revoking === invitation.id ? (
                      <div className="flex gap-2">
                        <Button
                          variant="destructive"
                          disabled={revoke.isPending}
                          onClick={() => revoke.mutate(invitation.id)}
                        >
                          Confirm revoke
                        </Button>
                        <Button
                          variant="outline"
                          onClick={() => setRevoking(null)}
                        >
                          Cancel
                        </Button>
                      </div>
                    ) : (
                      <Button
                        variant="outline"
                        onClick={() => setRevoking(invitation.id)}
                      >
                        Revoke…
                      </Button>
                    )}
                  </li>
                ))}
              </ul>
            ) : (
              <p className="text-sm text-muted-foreground">
                No pending invitations.
              </p>
            )}
            {revoke.isError && <p role="alert">{errorMessage(revoke.error)}</p>}
          </section>
        </>
      )}
    </div>
  );
}
