export default function ServicesLoading() {
  return (
    <div className="mx-auto max-w-7xl px-6 py-8">
      {/* Header skeleton */}
      <div className="mb-4 flex items-center justify-between">
        <div>
          <div className="bg-muted h-5 w-20 animate-pulse rounded" />
          <div className="bg-muted mt-1 h-4 w-56 animate-pulse rounded" />
        </div>
        <div className="bg-muted h-9 w-36 animate-pulse rounded" />
      </div>

      {/* Table skeleton */}
      <div className="bg-card rounded-lg border">
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="bg-muted/50 border-b">
                <th className="px-6 py-3 text-left">
                  <div className="bg-muted h-3 w-12 animate-pulse rounded" />
                </th>
                <th className="px-6 py-3 text-left">
                  <div className="bg-muted h-3 w-8 animate-pulse rounded" />
                </th>
                <th className="px-6 py-3 text-left">
                  <div className="bg-muted h-3 w-6 animate-pulse rounded" />
                </th>
                <th className="px-6 py-3 text-left">
                  <div className="bg-muted h-3 w-12 animate-pulse rounded" />
                </th>
                <th className="px-6 py-3 text-right">
                  <div className="bg-muted ml-auto h-3 w-16 animate-pulse rounded" />
                </th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {Array.from({ length: 5 }).map((_, i) => (
                <tr key={i}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center gap-2">
                      <div className="bg-muted h-5 w-5 animate-pulse rounded" />
                      <div className="bg-muted h-4 w-24 animate-pulse rounded" />
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="bg-muted h-4 w-40 animate-pulse rounded" />
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="bg-muted h-4 w-24 animate-pulse rounded" />
                  </td>
                  <td className="px-6 py-4">
                    <div className="bg-muted h-4 w-16 animate-pulse rounded" />
                  </td>
                  <td className="px-6 py-4 text-right">
                    <div className="bg-muted ml-auto h-8 w-16 animate-pulse rounded" />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
