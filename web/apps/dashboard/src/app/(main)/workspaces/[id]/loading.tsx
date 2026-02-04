export default function WorkspaceLoading() {
  return (
    <>
      {/* Back button skeleton */}
      <div className="mb-8">
        <div className="flex items-center gap-2">
          <div className="bg-muted h-4 w-4 animate-pulse rounded" />
          <div className="bg-muted h-4 w-32 animate-pulse rounded" />
        </div>
      </div>

      {/* Header skeleton */}
      <div className="mb-6">
        <div className="flex items-center justify-between gap-4 mb-2">
          <div className="bg-muted h-8 w-64 animate-pulse rounded" />
          <div className="flex gap-2">
            <div className="bg-muted h-9 w-24 animate-pulse rounded" />
            <div className="bg-muted h-9 w-9 animate-pulse rounded" />
          </div>
        </div>
        <div className="flex items-center gap-3">
          <div className="bg-muted h-4 w-20 animate-pulse rounded" />
          <div className="bg-muted h-4 w-4 animate-pulse rounded-full" />
          <div className="bg-muted h-4 w-24 animate-pulse rounded" />
          <div className="bg-muted h-4 w-4 animate-pulse rounded-full" />
          <div className="bg-muted h-6 w-20 animate-pulse rounded-full" />
        </div>
      </div>

      {/* Nav skeleton */}
      <div className="mb-5 pb-0 border-b">
        <div className="inline-flex gap-1">
          <div className="bg-muted h-9 w-20 animate-pulse rounded" />
          <div className="bg-muted h-9 w-20 animate-pulse rounded" />
          <div className="bg-muted h-9 w-20 animate-pulse rounded" />
          <div className="bg-muted h-9 w-24 animate-pulse rounded" />
          <div className="bg-muted h-9 w-20 animate-pulse rounded" />
        </div>
      </div>

      {/* Content skeleton */}
      <div className="space-y-6">
        <div className="bg-card rounded-lg border p-6">
          <div className="bg-muted mb-4 h-6 w-40 animate-pulse rounded" />
          <div className="grid gap-4 sm:grid-cols-2">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 rounded-lg border p-4">
                <div className="bg-muted h-10 w-10 animate-pulse rounded-lg" />
                <div className="flex-1">
                  <div className="bg-muted h-4 w-24 animate-pulse rounded" />
                  <div className="bg-muted mt-1 h-3 w-32 animate-pulse rounded" />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </>
  )
}
