export default function EnvironmentsLoading() {
  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Header skeleton */}
      <div className="mb-8">
        <div className="mb-6">
          <div className="bg-muted h-8 w-40 animate-pulse rounded" />
          <div className="bg-muted mt-1.5 h-4 w-80 animate-pulse rounded" />
        </div>

        <div className="space-y-4">
          {/* Filter and Actions */}
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              {/* Scope Filter */}
              <div className="bg-muted flex h-9 w-24 animate-pulse items-center gap-1 rounded-md" />
              {/* Status Filter */}
              <div className="bg-muted flex h-9 w-24 animate-pulse items-center gap-1 rounded-md" />
              {/* Count */}
              <div className="bg-muted h-4 w-24 animate-pulse rounded" />
            </div>
            <div className="flex items-center gap-2">
              <div className="bg-muted h-9 w-20 animate-pulse rounded" />
              <div className="bg-muted h-9 w-36 animate-pulse rounded" />
            </div>
          </div>

          {/* Table skeleton */}
          <div className="bg-card overflow-hidden rounded-lg border">
            <table className="min-w-full">
              <thead className="bg-muted/50 border-b">
                <tr>
                  <th className="px-6 py-3 text-left">
                    <div className="bg-muted h-3 w-12 animate-pulse rounded" />
                  </th>
                  <th className="px-6 py-3 text-left">
                    <div className="bg-muted h-3 w-14 animate-pulse rounded" />
                  </th>
                  <th className="px-6 py-3 text-center">
                    <div className="bg-muted mx-auto h-3 w-16 animate-pulse rounded" />
                  </th>
                  <th className="px-6 py-3 text-left">
                    <div className="bg-muted h-3 w-14 animate-pulse rounded" />
                  </th>
                  <th className="px-6 py-3 text-right">
                    <div className="bg-muted ml-auto h-3 w-16 animate-pulse rounded" />
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {Array.from({ length: 6 }).map((_, i) => (
                  <tr key={i}>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="bg-muted h-4 w-40 animate-pulse rounded" />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="bg-muted h-4 w-24 animate-pulse rounded" />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="bg-muted mx-auto h-5 w-16 animate-pulse rounded-full" />
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="bg-muted h-5 w-16 animate-pulse rounded-full" />
                    </td>
                    <td className="px-6 py-4 text-right whitespace-nowrap">
                      <div className="bg-muted ml-auto h-8 w-8 animate-pulse rounded" />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </main>
  )
}
