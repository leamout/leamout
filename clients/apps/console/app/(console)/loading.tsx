import { Skeleton } from "@leamout/ui/components/skeleton";

export default function Loading() {
  return (
    <div role="status" aria-label="Loading page" className="space-y-8">
      <span className="sr-only">Loading page</span>
      <div aria-hidden="true" className="space-y-3">
        <Skeleton className="h-9 w-48" />
        <Skeleton className="h-5 w-72 max-w-full" />
      </div>
      <div aria-hidden="true" className="grid gap-4 md:grid-cols-2">
        {["first", "second", "third", "fourth"].map((key) => (
          <Skeleton key={key} className="h-36 rounded-xl" />
        ))}
      </div>
    </div>
  );
}
