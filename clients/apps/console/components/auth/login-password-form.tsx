import { Input } from "@leamout/ui/components/input";
import Link from "next/link";
import { AuthForm } from "./auth-form";

export function LoginPasswordForm() {
  return (
    <AuthForm
      title="Enter your password"
      description="Enter your password to continue."
      footer={
        <>
          <p>
            <Link href="/email-verification" className="underline">
              Use an email code instead
            </Link>
          </p>
          <p>
            <Link href="/forgot-password" className="underline">
              Forgot password?
            </Link>
          </p>
        </>
      }
    >
      <div className="space-y-2">
        <label htmlFor="password">Password</label>
        <Input
          id="password"
          type="password"
          autoComplete="current-password"
          required
        />
      </div>
    </AuthForm>
  );
}
