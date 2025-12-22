export default function SettingsLoading() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      <div className="flex gap-6">
        {/* Side nav skeleton */}
        <div className="w-48 space-y-1">
          {['General', 'Network', 'Security', 'Access', 'Danger Zone'].map((_, i) => (
            <div key={i} className="bg-muted h-9 w-full animate-pulse rounded" />
          ))}
        </div>

        {/* Content skeleton */}
        <div className="flex-1 space-y-6">
          {/* Section title */}
          <div>
            <div className="bg-muted h-7 w-32 animate-pulse rounded" />
            <div className="bg-muted mt-1 h-4 w-64 animate-pulse rounded" />
          </div>

          {/* Form fields skeleton */}
          <div className="bg-card rounded-lg border p-6">
            <div className="space-y-4">
              {Array.from({ length: 3 }).map((_, i) => (
                <div key={i} className="space-y-2">
                  <div className="bg-muted h-4 w-24 animate-pulse rounded" />
                  <div className="bg-muted h-10 w-full animate-pulse rounded" />
                </div>
              ))}
            </div>
            <div className="mt-6 flex justify-end">
              <div className="bg-muted h-10 w-24 animate-pulse rounded" />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
