"use client";

import { Button } from "@leamout/ui/components/button";
import { useState } from "react";

export function AuthForm({
  title,
  description,
  children,
  footer,
  submitLabel = "Continue",
}: {
  title: string;
  description: string;
  children: React.ReactNode;
  footer: React.ReactNode;
  submitLabel?: string;
}) {
  const [notice, setNotice] = useState("");

  return (
    <div className="space-y-6">
      <div className="space-y-2 text-center">
        <h1 className="text-3xl font-semibold tracking-tight">{title}</h1>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
      <form
        className="space-y-4"
        onSubmit={(event) => {
          event.preventDefault();
          setNotice(
            "This form is a preview. Authentication will be connected later.",
          );
        }}
      >
        {children}
        <Button className="w-full" type="submit">
          {submitLabel}
        </Button>
        <p role="status" className="text-sm text-muted-foreground">
          {notice}
        </p>
      </form>
      <div className="space-y-3 text-center text-sm">{footer}</div>
    </div>
  );
}
