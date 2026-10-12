import { MarketingFooter } from "@/components/marketing/footer";
import { MarketingHeader } from "@/components/marketing/header";

export default function MarketingLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const consoleURL = (
    process.env.CONSOLE_URL ??
    (process.env.NODE_ENV === "development"
      ? "http://localhost:3001"
      : "https://console.leamout.com")
  ).replace(/\/+$/, "");

  return (
    <div className="flex min-h-screen flex-col">
      <MarketingHeader consoleURL={consoleURL} />
      <main id="main-content" tabIndex={-1} className="flex-1 outline-none">
        {children}
      </main>
      <MarketingFooter />
    </div>
  );
}
