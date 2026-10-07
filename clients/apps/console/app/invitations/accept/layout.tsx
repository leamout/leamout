import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Accept invitation | Leamout",
  referrer: "no-referrer",
  robots: { index: false, follow: false },
};

export default function InvitationLayout({
  children,
}: {
  children: ReactNode;
}) {
  return children;
}
