export default function WorkspaceDetailLoading() {
  return (
    <>
      {/* Header skeleton */}
      <div className="bg-card border-b">
        <div className="mx-auto max-w-7xl px-6">
          {/* Breadcrumb skeleton */}
          <div className="py-4">
            <div className="flex items-center gap-2">
              <div className="bg-muted h-4 w-24 animate-pulse rounded" />
              <div className="bg-muted h-4 w-4 animate-pulse rounded" />
              <div className="bg-muted h-4 w-32 animate-pulse rounded" />
            </div>
          </div>

          {/* Title and actions skeleton */}
          <div className="pb-6">
            <div className="flex items-start justify-between">
              <div>
                <div className="bg-muted h-8 w-64 animate-pulse rounded" />
                <div className="bg-muted mt-2 h-4 w-48 animate-pulse rounded" />
                <div className="mt-3 flex items-center gap-4">
                  <div className="bg-muted h-6 w-20 animate-pulse rounded-full" />
                </div>
              </div>
              <div className="flex gap-2">
                <div className="bg-muted h-9 w-24 animate-pulse rounded" />
                <div className="bg-muted h-9 w-9 animate-pulse rounded" />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Main content skeleton */}
      <div className="mx-auto max-w-7xl px-6 py-8">
        <div className="grid gap-6 lg:grid-cols-3">
          {/* Connection options skeleton - 2/3 width */}
          <div className="lg:col-span-2">
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

          {/* Sidebar skeleton - 1/3 width */}
          <div className="space-y-6">
            {/* Packages card skeleton */}
            <div className="bg-card rounded-lg border">
              <div className="border-b p-4">
                <div className="flex items-center gap-2">
                  <div className="bg-muted h-8 w-8 animate-pulse rounded-lg" />
                  <div>
                    <div className="bg-muted h-4 w-20 animate-pulse rounded" />
                    <div className="bg-muted mt-1 h-3 w-28 animate-pulse rounded" />
                  </div>
                </div>
              </div>
              <div className="p-4">
                <div className="bg-muted h-10 w-full animate-pulse rounded" />
              </div>
            </div>

            {/* Metrics card skeleton */}
            <div className="bg-card rounded-lg border p-4">
              <div className="bg-muted mb-4 h-5 w-32 animate-pulse rounded" />
              <div className="space-y-4">
                <div>
                  <div className="mb-2 flex justify-between">
                    <div className="bg-muted h-3 w-12 animate-pulse rounded" />
                    <div className="bg-muted h-3 w-8 animate-pulse rounded" />
                  </div>
                  <div className="bg-muted h-2 w-full animate-pulse rounded-full" />
                </div>
                <div>
                  <div className="mb-2 flex justify-between">
                    <div className="bg-muted h-3 w-16 animate-pulse rounded" />
                    <div className="bg-muted h-3 w-12 animate-pulse rounded" />
                  </div>
                  <div className="bg-muted h-2 w-full animate-pulse rounded-full" />
                </div>
              </div>
            </div>

            {/* Info card skeleton */}
            <div className="bg-card rounded-lg border p-6">
              <div className="bg-muted mb-4 h-4 w-24 animate-pulse rounded" />
              <div className="space-y-3">
                {Array.from({ length: 4 }).map((_, i) => (
                  <div key={i} className="flex justify-between">
                    <div className="bg-muted h-4 w-20 animate-pulse rounded" />
                    <div className="bg-muted h-4 w-24 animate-pulse rounded" />
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  )
}
