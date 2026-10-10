import Link from "next/link";
import {
  type ProviderAvailability,
  type ProviderCredential,
  providerCatalog,
} from "@/components/ai-providers/catalog";
import { ProviderCredentialForm } from "@/components/ai-providers/provider-credential-form";

export function ProviderDetails({
  providerId,
  availability,
  credentials,
}: {
  providerId: string;
  availability?: ProviderAvailability;
  credentials?: ProviderCredential[];
}) {
  const provider = providerCatalog.find((item) => item.id === providerId);
  if (!provider)
    return (
      <div className="space-y-4">
        <h1 className="text-3xl font-semibold">Unknown provider</h1>
        <p className="text-muted-foreground">
          This provider is not in the console preview catalog.
        </p>
        <Link href="/ai-providers" className="underline">
          Back to AI providers
        </Link>
      </div>
    );
  return (
    <div className="space-y-8">
      <Link
        href="/ai-providers"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to AI providers
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          {provider.name}
        </h1>
        <p className="text-muted-foreground">
          Manage organization BYOA credentials for this provider.
        </p>
        <p className="text-sm text-muted-foreground">
          UI preview. Live provider data is not connected.
        </p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <section className="space-y-2 rounded-lg border p-6">
          <h2 className="font-semibold">Capabilities</h2>
          <p className="text-sm text-muted-foreground">
            {availability?.capabilities
              ? availability.capabilities.join(", ") || "None available"
              : "Provider capabilities have not been loaded."}
          </p>
        </section>
        <section className="space-y-2 rounded-lg border p-6">
          <h2 className="font-semibold">Platform default</h2>
          <p className="text-sm text-muted-foreground">
            {availability?.platformAvailable === undefined
              ? "Availability has not been loaded."
              : availability.platformAvailable
                ? "A platform credential is available."
                : "No platform credential is available. Configure organization BYOA credentials."}
          </p>
          <p className="text-sm text-muted-foreground">
            Platform credentials are configured by the deployment operator.
            Their secrets are never displayed here.
          </p>
        </section>
      </div>
      <section className="space-y-4">
        <h2 className="text-lg font-semibold">Organization credentials</h2>
        {credentials?.length ? (
          <ul className="space-y-3">
            {credentials.map((credential) => (
              <li key={credential.id} className="rounded-lg border p-4">
                <p className="font-medium">{credential.name}</p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Connection: {credential.connectionState}
                </p>
                <p className="mt-1 text-sm text-muted-foreground">
                  Verified: {credential.verifiedAt ?? "Not verified"}
                </p>
              </li>
            ))}
          </ul>
        ) : (
          <p className="rounded-lg border border-dashed p-6 text-sm text-muted-foreground">
            {credentials
              ? "No organization credentials configured."
              : "Organization credentials have not been loaded."}
          </p>
        )}
      </section>
      <ProviderCredentialForm />
      <p className="text-sm text-muted-foreground">
        Assign organization credentials or platform defaults from{" "}
        <Link href="/agents" className="underline">
          agent configuration
        </Link>
        .
      </p>
    </div>
  );
}
