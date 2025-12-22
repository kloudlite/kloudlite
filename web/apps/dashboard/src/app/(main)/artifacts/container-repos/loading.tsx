export default function ContainerReposLoading() {
  return (
    <div className="space-y-4">
      {/* Header skeleton */}
      <div className="flex items-center justify-between">
        <div className="bg-muted h-4 w-28 animate-pulse rounded" />
      </div>

      {/* Repository list skeleton */}
      <div className="rounded-lg border">
        {Array.from({ length: 5 }).map((_, i) => (
          <div key={i} className="flex items-center justify-between border-b px-4 py-3 last:border-b-0">
            <div className="flex items-center gap-3">
              <div className="bg-muted h-10 w-10 animate-pulse rounded" />
              <div>
                <div className="bg-muted h-5 w-40 animate-pulse rounded" />
                <div className="bg-muted mt-1 h-3 w-24 animate-pulse rounded" />
              </div>
            </div>
            <div className="bg-muted h-8 w-8 animate-pulse rounded" />
          </div>
        ))}
      </div>
    </div>
  )
}
