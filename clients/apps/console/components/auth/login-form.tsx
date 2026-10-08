import { Input } from "@leamout/ui/components/input";
import Link from "next/link";
import { AuthForm } from "./auth-form";

export function LoginForm() {
  return (
    <AuthForm
      title="Log in to Leamout"
      description="Continue with your email address."
      footer={
        <p>
          New to Leamout?{" "}
          <Link href="/create-account" className="underline">
            Create account
          </Link>
        </p>
      }
    >
      <div className="space-y-2">
        <label htmlFor="email">Email address</label>
        <Input id="email" type="email" autoComplete="email" required />
      </div>
    </AuthForm>
  );
}
