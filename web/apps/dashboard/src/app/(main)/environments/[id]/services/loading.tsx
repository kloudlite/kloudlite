export default function ServicesLoading() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Header skeleton */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <div className="bg-muted h-6 w-24 animate-pulse rounded" />
          <div className="bg-muted mt-1 h-4 w-64 animate-pulse rounded" />
        </div>
        <div className="bg-muted h-9 w-28 animate-pulse rounded" />
      </div>

      {/* Services grid skeleton */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="bg-card rounded-lg border p-4">
            <div className="mb-3 flex items-center justify-between">
              <div className="flex items-center gap-2">
                <div className="bg-muted h-8 w-8 animate-pulse rounded" />
                <div className="bg-muted h-5 w-28 animate-pulse rounded" />
              </div>
              <div className="bg-muted h-5 w-16 animate-pulse rounded-full" />
            </div>
            <div className="space-y-2">
              <div className="bg-muted h-4 w-full animate-pulse rounded" />
              <div className="bg-muted h-4 w-3/4 animate-pulse rounded" />
            </div>
            <div className="mt-4 flex gap-2">
              <div className="bg-muted h-8 w-20 animate-pulse rounded" />
              <div className="bg-muted h-8 w-20 animate-pulse rounded" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
