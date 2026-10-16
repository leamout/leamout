import { IdentityEntry } from "@/components/auth/identity-entry";

export default function Page() {
  return (
    <IdentityEntry
      title="Welcome to Leamout"
      description="Sign in to manage your voice agents."
      register={false}
    />
  );
}
