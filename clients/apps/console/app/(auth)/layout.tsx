import Link from "next/link";
import { AuthFlowProvider } from "@/components/auth/auth-flow-provider";

export default function AuthLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <AuthFlowProvider>
      <div className="flex min-h-svh flex-col items-center px-6 py-10">
        <Link href="/log-in" className="text-xl font-semibold tracking-tight">
          Leamout
        </Link>
        <main className="flex w-full max-w-sm flex-1 flex-col justify-center py-12">
          {children}
        </main>
      </div>
    </AuthFlowProvider>
  );
}
