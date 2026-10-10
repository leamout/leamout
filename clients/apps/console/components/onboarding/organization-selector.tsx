"use client";

import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { useState } from "react";

type Organization = { id: string; name: string };

export function OrganizationSelector({
  organizations = [],
}: {
  organizations?: Organization[];
}) {
  const [notice, setNotice] = useState("");

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-semibold tracking-tight">
          Select an organization
        </h1>
        <p className="text-sm text-muted-foreground">
          Choose the workspace you want to manage.
        </p>
      </div>
      {organizations.length > 0 ? (
        <ul className="space-y-3" aria-label="Organizations">
          {organizations.map((organization) => (
            <li key={organization.id}>
              <Button
                className="h-auto w-full justify-start whitespace-normal break-words py-4 text-left"
                variant="outline"
                onClick={() =>
                  setNotice(
                    "This is a preview. Organization switching will be connected later.",
                  )
                }
              >
                {organization.name}
              </Button>
            </li>
          ))}
        </ul>
      ) : (
        <div className="rounded-xl border border-dashed p-6 text-center">
          <p className="font-medium">No organizations to display</p>
          <p className="mt-2 text-sm text-muted-foreground">
            Your organizations will appear here once account integration is
            connected.
          </p>
        </div>
      )}
      <Button className="w-full" render={<Link href="/create-organization" />}>
        Create organization
      </Button>
      <p role="status" className="text-sm text-muted-foreground">
        {notice}
      </p>
      <p className="text-center text-sm">
        <Link href="/log-in" className="underline">
          Back to log in
        </Link>
      </p>
    </div>
  );
}
