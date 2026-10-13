import { Button } from "@leamout/ui/components/button";
import Link from "next/link";
import { consoleURL } from "@/lib/site";

export function AboutCallToAction() {
  return (
    <div className="py-16">
      <div className="flex flex-wrap gap-3">
        <Button
          size="lg"
          nativeButton={false}
          render={<a href={`${consoleURL}/create-account`} />}
        >
          Get started
        </Button>
        <Button
          size="lg"
          variant="outline"
          nativeButton={false}
          render={<Link href="/docs" />}
        >
          Read documentation
        </Button>
      </div>
    </div>
  );
}
