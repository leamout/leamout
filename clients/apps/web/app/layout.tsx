import "@leamout/ui/globals.css";
import { plexMono, plexSans } from "@/lib/fonts";
import { constructMetadata } from "@/lib/metadata";
import RootProviders from "./providers";

export const metadata = constructMetadata({
  title: "Carrier-grade runtime for autonomous voice agents",
  description:
    "Leamout is the carrier-grade runtime and control plane for autonomous voice agents.",
});

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
      <body className="bg-background text-foreground antialiased">
        <RootProviders>{children}</RootProviders>
      </body>
    </html>
  );
}
