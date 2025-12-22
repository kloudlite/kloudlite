export default function EnvironmentDetailLoading() {
  return (
    <>
      {/* Header skeleton */}
      <div className="bg-background border-b">
        <div className="mx-auto max-w-7xl px-6">
          {/* Breadcrumb skeleton */}
          <div className="py-4">
            <div className="flex items-center gap-2">
              <div className="bg-muted h-4 w-28 animate-pulse rounded" />
              <div className="bg-muted h-4 w-4 animate-pulse rounded" />
              <div className="bg-muted h-4 w-36 animate-pulse rounded" />
            </div>
          </div>

          {/* Title skeleton */}
          <div className="pb-4">
            <div className="flex items-start justify-between">
              <div>
                <div className="bg-muted h-8 w-56 animate-pulse rounded" />
                <div className="mt-1.5 flex items-center gap-4">
                  <div className="bg-muted h-4 w-32 animate-pulse rounded" />
                  <div className="bg-muted h-4 w-4 animate-pulse rounded-full" />
                  <div className="bg-muted h-4 w-28 animate-pulse rounded" />
                  <div className="bg-muted h-4 w-4 animate-pulse rounded-full" />
                  <div className="bg-muted h-6 w-16 animate-pulse rounded-full" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Navigation skeleton */}
      <div className="border-b bg-background">
        <div className="mx-auto max-w-7xl px-6">
          <div className="flex gap-6 py-3">
            {['Services', 'Configs', 'Settings'].map((_, i) => (
              <div key={i} className="bg-muted h-5 w-20 animate-pulse rounded" />
            ))}
          </div>
        </div>
      </div>

      {/* Content skeleton */}
      <div className="mx-auto max-w-7xl px-6 py-8">
        <div className="space-y-4">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="bg-card rounded-lg border p-6">
              <div className="bg-muted mb-4 h-5 w-40 animate-pulse rounded" />
              <div className="bg-muted h-20 w-full animate-pulse rounded" />
            </div>
          ))}
        </div>
      </div>
    </>
  )
}
