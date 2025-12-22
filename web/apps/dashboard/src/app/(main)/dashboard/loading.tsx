export default function DashboardLoading() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Header skeleton */}
      <div className="mb-8">
        <div className="bg-muted h-8 w-48 animate-pulse rounded" />
        <div className="bg-muted mt-2 h-4 w-64 animate-pulse rounded" />
      </div>

      {/* Stats grid skeleton */}
      <div className="mb-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="bg-card rounded-lg border p-6">
            <div className="bg-muted h-4 w-24 animate-pulse rounded" />
            <div className="bg-muted mt-2 h-8 w-16 animate-pulse rounded" />
          </div>
        ))}
      </div>

      {/* Content skeleton */}
      <div className="grid gap-6 lg:grid-cols-2">
        {Array.from({ length: 2 }).map((_, i) => (
          <div key={i} className="bg-card rounded-lg border p-6">
            <div className="bg-muted mb-4 h-5 w-32 animate-pulse rounded" />
            <div className="space-y-3">
              {Array.from({ length: 3 }).map((_, j) => (
                <div key={j} className="bg-muted h-12 animate-pulse rounded" />
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
