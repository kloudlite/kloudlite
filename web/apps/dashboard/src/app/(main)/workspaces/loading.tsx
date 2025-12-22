export default function WorkspacesLoading() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Header skeleton */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <div className="bg-muted h-8 w-36 animate-pulse rounded" />
          <div className="bg-muted mt-2 h-4 w-56 animate-pulse rounded" />
        </div>
        <div className="bg-muted h-10 w-36 animate-pulse rounded" />
      </div>

      {/* Workspace cards skeleton */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <div key={i} className="bg-card rounded-lg border p-6">
            <div className="mb-4 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="bg-muted h-10 w-10 animate-pulse rounded-lg" />
                <div>
                  <div className="bg-muted h-5 w-28 animate-pulse rounded" />
                  <div className="bg-muted mt-1 h-3 w-20 animate-pulse rounded" />
                </div>
              </div>
              <div className="bg-muted h-6 w-16 animate-pulse rounded-full" />
            </div>
            <div className="space-y-2">
              <div className="bg-muted h-4 w-full animate-pulse rounded" />
              <div className="bg-muted h-4 w-2/3 animate-pulse rounded" />
            </div>
            <div className="mt-4 flex gap-2">
              <div className="bg-muted h-8 flex-1 animate-pulse rounded" />
              <div className="bg-muted h-8 w-8 animate-pulse rounded" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
