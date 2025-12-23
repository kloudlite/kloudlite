export default function CodeAnalysisLoading() {
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
              <div className="bg-muted h-4 w-28 animate-pulse rounded" />
              <div className="bg-muted h-4 w-4 animate-pulse rounded" />
              <div className="bg-muted h-4 w-28 animate-pulse rounded" />
            </div>
          </div>

          {/* Title and actions skeleton */}
          <div className="flex items-center justify-between pb-6">
            <div>
              <div className="flex items-center gap-3">
                <div className="bg-muted h-5 w-5 animate-pulse rounded" />
                <div className="bg-muted h-8 w-40 animate-pulse rounded" />
              </div>
              <div className="bg-muted ml-8 mt-2 h-4 w-48 animate-pulse rounded" />
            </div>
            <div className="bg-muted h-10 w-28 animate-pulse rounded" />
          </div>
        </div>
      </div>

      {/* Main content skeleton */}
      <div className="mx-auto max-w-7xl px-6 py-8">
        {/* Summary cards skeleton */}
        <div className="mb-8 grid gap-4 md:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="bg-card rounded-lg border p-4">
              <div className="flex items-center gap-3">
                <div className="bg-muted h-10 w-10 animate-pulse rounded-lg" />
                <div>
                  <div className="bg-muted h-5 w-28 animate-pulse rounded" />
                  <div className="bg-muted mt-1 h-4 w-20 animate-pulse rounded" />
                </div>
              </div>
              <div className="mt-4 grid grid-cols-4 gap-2">
                {Array.from({ length: 4 }).map((_, j) => (
                  <div key={j} className="text-center">
                    <div className="bg-muted mx-auto h-6 w-6 animate-pulse rounded" />
                    <div className="bg-muted mx-auto mt-1 h-3 w-10 animate-pulse rounded" />
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Filters skeleton */}
        <div className="mb-6 flex flex-wrap items-center gap-4">
          <div className="bg-muted h-9 w-48 animate-pulse rounded-md" />
          <div className="bg-muted h-9 w-72 animate-pulse rounded-md" />
          <div className="bg-muted h-9 flex-1 min-w-[200px] animate-pulse rounded-md" />
        </div>

        {/* Table skeleton */}
        <div className="bg-card overflow-hidden rounded-lg border">
          {/* Table header */}
          <div className="bg-muted/50 border-b px-4 py-3">
            <div className="flex items-center gap-4">
              <div className="bg-muted h-4 w-10 animate-pulse rounded" />
              <div className="bg-muted h-4 w-16 animate-pulse rounded" />
              <div className="bg-muted h-4 w-20 animate-pulse rounded" />
              <div className="bg-muted h-4 w-20 animate-pulse rounded" />
              <div className="bg-muted h-4 w-16 animate-pulse rounded" />
            </div>
          </div>

          {/* Table rows */}
          {Array.from({ length: 8 }).map((_, i) => (
            <div key={i} className="border-b px-4 py-3">
              <div className="flex items-center gap-4">
                <div className="bg-muted h-4 w-4 animate-pulse rounded" />
                <div className="flex items-center gap-2 flex-1">
                  <div className="bg-muted h-4 w-4 animate-pulse rounded" />
                  <div className="bg-muted h-4 w-48 animate-pulse rounded" />
                </div>
                <div className="bg-muted h-4 w-20 animate-pulse rounded" />
                <div className="bg-muted h-4 w-32 animate-pulse rounded" />
                <div className="bg-muted h-5 w-16 animate-pulse rounded-full" />
              </div>
            </div>
          ))}

          {/* Table footer */}
          <div className="border-t px-4 py-3">
            <div className="bg-muted h-4 w-32 animate-pulse rounded" />
          </div>
        </div>
      </div>
    </>
  )
}
