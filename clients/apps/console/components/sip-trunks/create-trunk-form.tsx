import Link from "next/link";
import { TrunkForm } from "@/components/sip-trunks/trunk-form";

export function CreateTrunkForm() {
  return (
    <div className="space-y-6">
      <Link
        href="/sip-trunks"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to SIP trunks
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Create SIP trunk
        </h1>
        <p className="text-muted-foreground">
          Define your BYOC connection, then configure endpoints and
          authentication.
        </p>
      </div>
      <TrunkForm creating />
    </div>
  );
}
