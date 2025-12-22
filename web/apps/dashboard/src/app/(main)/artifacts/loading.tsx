export default function ArtifactsLoading() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Header skeleton */}
      <div className="mb-6 flex items-center justify-between">
        <div>
          <div className="bg-muted h-8 w-32 animate-pulse rounded" />
          <div className="bg-muted mt-2 h-4 w-48 animate-pulse rounded" />
        </div>
        <div className="bg-muted h-10 w-32 animate-pulse rounded" />
      </div>

      {/* Artifacts list skeleton */}
      <div className="bg-card rounded-lg border">
        <div className="border-b p-4">
          <div className="bg-muted h-5 w-24 animate-pulse rounded" />
        </div>
        <div className="divide-y">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center justify-between p-4">
              <div className="flex items-center gap-4">
                <div className="bg-muted h-10 w-10 animate-pulse rounded" />
                <div>
                  <div className="bg-muted h-5 w-40 animate-pulse rounded" />
                  <div className="bg-muted mt-1 h-3 w-24 animate-pulse rounded" />
                </div>
              </div>
              <div className="flex items-center gap-2">
                <div className="bg-muted h-8 w-8 animate-pulse rounded" />
                <div className="bg-muted h-8 w-8 animate-pulse rounded" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
