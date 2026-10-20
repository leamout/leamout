import { MarketingFooter } from "@/components/marketing/footer";
import { MarketingHeader } from "@/components/marketing/header";
import { consoleURL } from "@/lib/site";
import "./marketing.css";

export default function MarketingLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <div className="leamout-marketing dark flex min-h-screen flex-col bg-background text-foreground">
      <MarketingHeader consoleURL={consoleURL} />
      <main id="main-content" tabIndex={-1} className="flex-1 outline-none">
        {children}
      </main>
      <MarketingFooter />
    </div>
  );
}
