import { Input } from "@leamout/ui/components/input";
import Link from "next/link";
import { AuthForm } from "./auth-form";

export function EmailVerificationForm() {
  return (
    <AuthForm
      title="Check your email"
      description="Enter the six-digit verification code."
      footer={
        <>
          <p className="text-muted-foreground">
            Resending codes will be available once authentication is connected.
          </p>
          <Link href="/log-in" className="underline">
            Change email
          </Link>
        </>
      }
    >
      <div className="space-y-2">
        <label htmlFor="code">Verification code</label>
        <Input
          id="code"
          inputMode="numeric"
          autoComplete="one-time-code"
          pattern="[0-9]{6}"
          maxLength={6}
          required
        />
      </div>
    </AuthForm>
  );
}
