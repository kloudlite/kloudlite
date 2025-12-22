export default function Loading() {
  return (
    <main className="min-h-screen">
      <div className="mx-auto max-w-7xl px-6 py-8">
        {/* Page Header */}
        <div className="mb-8">
          <div className="bg-muted h-8 w-32 animate-pulse rounded" />
          <div className="bg-muted mt-1.5 h-4 w-72 animate-pulse rounded" />
        </div>

        {/* Machine Info Card */}
        <div className="bg-card mb-6 border p-6">
          <div className="mb-6 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="bg-muted h-10 w-10 animate-pulse" />
              <div>
                <div className="bg-muted h-5 w-40 animate-pulse rounded" />
                <div className="bg-muted mt-1 h-3 w-24 animate-pulse rounded" />
              </div>
            </div>
            {/* Controls skeleton */}
            <div className="flex items-center gap-2">
              <div className="bg-muted h-9 w-20 animate-pulse rounded" />
              <div className="bg-muted h-9 w-9 animate-pulse rounded" />
            </div>
          </div>

          {/* Machine Stats - 4 column grid */}
          <div className="grid grid-cols-4 gap-6 border-t pt-6">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i}>
                <div className="bg-muted h-3 w-16 animate-pulse rounded" />
                <div className="bg-muted mt-2 h-5 w-24 animate-pulse rounded" />
              </div>
            ))}
          </div>
        </div>

        {/* Resource Usage Section */}
        <div className="mb-6">
          <div className="bg-muted mb-4 h-5 w-32 animate-pulse rounded" />
          <div className="grid gap-4 md:grid-cols-3">
            {Array.from({ length: 3 }).map((_, i) => (
              <div key={i} className="bg-card border p-4">
                <div className="mb-2 flex items-center justify-between">
                  <div className="bg-muted h-4 w-16 animate-pulse rounded" />
                  <div className="bg-muted h-4 w-12 animate-pulse rounded" />
                </div>
                <div className="bg-muted h-2 w-full animate-pulse rounded-full" />
              </div>
            ))}
          </div>
        </div>

        {/* Quick Access Section */}
        <div>
          <div className="bg-muted mb-4 h-5 w-28 animate-pulse rounded" />
          <div className="grid gap-4 md:grid-cols-2">
            {Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="bg-card border p-4">
                <div className="bg-muted mb-3 h-4 w-24 animate-pulse rounded" />
                <div className="space-y-2">
                  {Array.from({ length: 2 }).map((_, j) => (
                    <div key={j} className="flex items-center gap-2">
                      <div className="bg-muted h-8 w-8 animate-pulse rounded" />
                      <div className="bg-muted h-4 w-32 animate-pulse rounded" />
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </main>
  )
}
