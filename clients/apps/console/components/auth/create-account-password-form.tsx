import { Input } from "@leamout/ui/components/input";
import Link from "next/link";
import { AuthForm } from "./auth-form";

export function CreateAccountPasswordForm() {
  return (
    <AuthForm
      title="Create a password"
      description="Enter your password to continue."
      footer={
        <Link href="/log-in" className="underline">
          Back to log in
        </Link>
      }
    >
      <div className="space-y-2">
        <label htmlFor="password">Password</label>
        <Input
          id="password"
          type="password"
          autoComplete="new-password"
          required
        />
      </div>
      <div className="space-y-2">
        <label htmlFor="confirmation">Confirm password</label>
        <Input
          id="confirmation"
          type="password"
          autoComplete="new-password"
          required
        />
      </div>
    </AuthForm>
  );
}
