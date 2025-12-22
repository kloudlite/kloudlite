export default function EnvVarsLoading() {
  return (
    <div className="space-y-4">
      {/* Header skeleton */}
      <div className="mb-4 flex items-center justify-between">
        <div>
          <div className="bg-muted h-6 w-24 animate-pulse rounded" />
          <div className="bg-muted mt-1 h-4 w-48 animate-pulse rounded" />
        </div>
        <div className="bg-muted h-9 w-28 animate-pulse rounded" />
      </div>

      {/* Table skeleton */}
      <div className="rounded-lg border">
        {/* Table header */}
        <div className="border-b bg-muted/30 px-4 py-3">
          <div className="grid grid-cols-4 gap-4">
            <div className="bg-muted h-4 w-12 animate-pulse rounded" />
            <div className="bg-muted h-4 w-16 animate-pulse rounded" />
            <div className="bg-muted h-4 w-10 animate-pulse rounded" />
            <div className="bg-muted h-4 w-16 animate-pulse rounded" />
          </div>
        </div>
        {/* Table rows */}
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="border-b px-4 py-3 last:border-b-0">
            <div className="grid grid-cols-4 gap-4">
              <div className="bg-muted h-4 w-28 animate-pulse rounded" />
              <div className="bg-muted h-4 w-40 animate-pulse rounded" />
              <div className="bg-muted h-5 w-14 animate-pulse rounded-full" />
              <div className="flex gap-2">
                <div className="bg-muted h-8 w-8 animate-pulse rounded" />
                <div className="bg-muted h-8 w-8 animate-pulse rounded" />
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
