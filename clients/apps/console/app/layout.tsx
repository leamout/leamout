import { Toaster } from "@leamout/ui/components/sonner";
import { TooltipProvider } from "@leamout/ui/components/tooltip";
import type { Metadata } from "next";
import { ConsoleShell } from "@/components/console-shell";
import "@leamout/ui/globals.css";
import { plexMono, plexSans } from "@/lib/fonts";

export const metadata: Metadata = {
  title: "Leamout",
  description: "Configure and monitor your Leamout voice agents",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      suppressHydrationWarning
      className={`${plexSans.variable} ${plexMono.variable}`}
    >
      <body className="flex min-h-full flex-col bg-background text-foreground antialiased">
        <TooltipProvider>
          <ConsoleShell>{children}</ConsoleShell>
          <Toaster richColors closeButton />
        </TooltipProvider>
      </body>
    </html>
  );
}
