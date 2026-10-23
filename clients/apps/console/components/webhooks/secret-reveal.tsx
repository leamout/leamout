"use client";

import { Button } from "@leamout/ui/components/button";
import { useState } from "react";

export function SecretReveal({ secret, label }: { secret: string; label: string }) {
  const [visible, setVisible] = useState(true);
  return (
    <section className="space-y-3 rounded-lg border p-5">
      <h3 className="font-semibold">{label}</h3>
      <p className="text-sm text-muted-foreground">
        Copy this secret now. It is shown only in this response and is not available
        from subsequent API reads. Do not put it in logs or screenshots.
      </p>
      {visible ? (
        <div className="flex flex-wrap items-center gap-3">
          <code className="max-w-full break-all rounded bg-muted p-3 text-xs">{secret}</code>
          <Button variant="outline" onClick={() => void navigator.clipboard.writeText(secret)}>
            Copy
          </Button>
          <Button variant="outline" onClick={() => setVisible(false)}>Hide</Button>
        </div>
      ) : (
        <p className="text-sm text-muted-foreground">Secret hidden.</p>
      )}
    </section>
  );
}
