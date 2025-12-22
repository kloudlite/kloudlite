export default function EnvironmentsLoading() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Header skeleton */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <div className="bg-muted h-8 w-40 animate-pulse rounded" />
          <div className="bg-muted mt-2 h-4 w-64 animate-pulse rounded" />
        </div>
        <div className="bg-muted h-10 w-40 animate-pulse rounded" />
      </div>

      {/* Environment cards skeleton */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="bg-card rounded-lg border p-6">
            <div className="mb-4 flex items-center justify-between">
              <div className="bg-muted h-5 w-32 animate-pulse rounded" />
              <div className="bg-muted h-6 w-16 animate-pulse rounded-full" />
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
