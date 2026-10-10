import { Skeleton } from "@leamout/ui/components/skeleton";

export default function Loading() {
  return (
    <div
      role="status"
      aria-label="Loading network policies"
      className="space-y-6"
    >
      <span className="sr-only">Loading network policies</span>
      <div aria-hidden="true" className="space-y-6">
        <Skeleton className="h-9 w-40" />
        <Skeleton className="h-5 w-80 max-w-full" />
        <Skeleton className="h-10 w-full" />
        <Skeleton className="h-72 w-full rounded-lg" />
      </div>
    </div>
  );
}
