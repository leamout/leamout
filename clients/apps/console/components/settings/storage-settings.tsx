"use client";

import { useState } from "react";
import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

export function StorageSettings() {
  const [pathStyle, setPathStyle] = useState(false);
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Storage</h1>
      <section className="space-y-2 rounded-lg border p-6">
        <h2 className="font-semibold">Platform default</h2>
        <p className="text-sm text-muted-foreground">
          Default storage availability has not been loaded. Deployment
          credentials are managed by the operator.
        </p>
      </section>
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Organization storage (BYOS)</h2>
        <p className="text-sm text-muted-foreground">
          Configure S3-compatible recording storage. Existing storage
          configurations have not been loaded.
        </p>
        <label className="flex items-center gap-3 text-sm">
          <input
            type="checkbox"
            checked={pathStyle}
            onChange={(event) => setPathStyle(event.target.checked)}
          />
          Use path-style addressing
        </label>
        <PreviewSettingsForm
          action="Save storage"
          fields={[
            {
              name: "storageName",
              label: "Configuration name",
              required: true,
            },
            {
              name: "endpoint",
              label: "Endpoint URL",
              type: "url",
              required: true,
            },
            { name: "region", label: "Region", required: true },
            { name: "bucket", label: "Bucket", required: true },
            {
              name: "accessKey",
              label: "Access key ID",
              type: "password",
              required: true,
            },
            {
              name: "secretKey",
              label: "Secret access key",
              type: "password",
              required: true,
            },
          ]}
        />
      </section>
    </div>
  );
}
