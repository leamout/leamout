import Link from "next/link";

export function IdentityEntry({
  title,
  description,
  register,
}: {
  title: string;
  description: string;
  register: boolean;
}) {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-semibold">{title}</h1>
        <p className="text-muted-foreground">{description}</p>
      </div>
      <a
        href={
          register ? "/api/v1/auth/login?register=true" : "/api/v1/auth/login"
        }
        className="flex h-10 w-full items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground"
      >
        {register ? "Create account" : "Continue to sign in"}
      </a>
      <p className="text-center text-sm text-muted-foreground">
        {register ? "Already have an account? " : "New to Leamout? "}
        <Link
          href={register ? "/log-in" : "/create-account"}
          className="underline"
        >
          {register ? "Log in" : "Create an account"}
        </Link>
      </p>
    </div>
  );
}
