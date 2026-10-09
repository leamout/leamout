import Link from "next/link";
import { PhoneNumberForm } from "@/components/phone-numbers/phone-number-form";

export function ConnectPhoneNumberForm() {
  return (
    <div className="space-y-6">
      <Link
        href="/phone-numbers"
        className="text-sm text-muted-foreground hover:underline"
      >
        Back to phone numbers
      </Link>
      <div className="space-y-2">
        <h1 className="text-3xl font-semibold tracking-tight">
          Connect phone number
        </h1>
        <p className="text-muted-foreground">
          Bring an existing number from your carrier or PBX. This does not
          purchase or provision a number.
        </p>
      </div>
      <PhoneNumberForm creating />
    </div>
  );
}
