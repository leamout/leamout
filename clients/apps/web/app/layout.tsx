import "@leamout/ui/globals.css";
import { Analytics } from "@vercel/analytics/next";
import { SpeedInsights } from "@vercel/speed-insights/next";
import { plexMono, plexSans, spaceGrotesk } from "@/lib/fonts";
import { constructMetadata } from "@/lib/metadata";

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
      className={`${plexSans.variable} ${plexMono.variable} ${spaceGrotesk.variable}`}
    >
      <body className="bg-background text-foreground antialiased">
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  );
}
