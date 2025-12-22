export default function ConfigFilesLoading() {
  return (
    <div className="space-y-4">
      {/* Header skeleton */}
      <div className="mb-4 flex items-center justify-between">
        <div>
          <div className="bg-muted h-6 w-32 animate-pulse rounded" />
          <div className="bg-muted mt-1 h-4 w-56 animate-pulse rounded" />
        </div>
        <div className="bg-muted h-9 w-24 animate-pulse rounded" />
      </div>

      {/* Files list skeleton */}
      <div className="rounded-lg border">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="flex items-center justify-between border-b px-4 py-3 last:border-b-0">
            <div className="flex items-center gap-3">
              <div className="bg-muted h-8 w-8 animate-pulse rounded" />
              <div>
                <div className="bg-muted h-4 w-32 animate-pulse rounded" />
                <div className="bg-muted mt-1 h-3 w-24 animate-pulse rounded" />
              </div>
            </div>
            <div className="flex gap-2">
              <div className="bg-muted h-8 w-8 animate-pulse rounded" />
              <div className="bg-muted h-8 w-8 animate-pulse rounded" />
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
