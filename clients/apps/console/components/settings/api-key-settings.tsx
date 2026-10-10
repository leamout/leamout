import { PreviewSettingsForm } from "@/components/settings/preview-settings-form";

type ApiKey = {
  id: string;
  name: string;
  prefix: string;
  scopes: string[];
  lastUsed?: string;
};

export function ApiKeySettings({ keys = [] }: { keys?: ApiKey[] }) {
  return (
    <div className="space-y-8">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">API keys</h1>
        <p className="text-muted-foreground">
          Manage organization API credentials. No token is generated in this
          preview.
        </p>
      </div>
      {keys.length ? (
        <ul className="space-y-3">
          {keys.map((key) => (
            <li key={key.id} className="space-y-1 rounded-lg border p-4">
              <h2 className="font-medium">{key.name}</h2>
              <p className="font-mono text-sm">{key.prefix}…</p>
              <p className="text-sm text-muted-foreground">
                Scopes: {key.scopes.join(", ")}
              </p>
              <p className="text-sm text-muted-foreground">
                Last used: {key.lastUsed ?? "Never"}
              </p>
            </li>
          ))}
        </ul>
      ) : (
        <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
          API keys have not been loaded.
        </p>
      )}
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Create API key</h2>
        <p className="text-sm text-muted-foreground">
          Scope selection and expiry will be configured when credential
          integration is connected.
        </p>
        <PreviewSettingsForm
          action="Create API key"
          fields={[
            { name: "keyName", label: "Key name", required: true },
            { name: "keyDescription", label: "Description", type: "textarea" },
          ]}
        />
      </section>
    </div>
  );
}
