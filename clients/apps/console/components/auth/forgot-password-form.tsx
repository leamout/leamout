import { Input } from "@leamout/ui/components/input";
import Link from "next/link";
import { AuthForm } from "./auth-form";

export function ForgotPasswordForm() {
  return (
    <AuthForm
      title="Forgot your password?"
      description="Continue with your email address."
      footer={
        <Link href="/log-in" className="underline">
          Back to log in
        </Link>
      }
    >
      <div className="space-y-2">
        <label htmlFor="email">Email address</label>
        <Input id="email" type="email" autoComplete="email" required />
      </div>
    </AuthForm>
  );
}
