import { IdentityEntry } from "@/components/auth/identity-entry";

export default function Page() {
  return (
    <IdentityEntry
      title="Create your account"
      description="Get started with Leamout."
      register={true}
    />
  );
}
