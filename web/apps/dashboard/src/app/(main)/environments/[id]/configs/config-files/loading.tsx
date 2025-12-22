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

      {/* Table skeleton */}
      <div className="bg-card overflow-hidden rounded-lg border">
        <table className="min-w-full">
          <thead className="bg-muted/50 border-b">
            <tr>
              <th className="text-muted-foreground px-6 py-3 text-left text-xs font-medium tracking-wider uppercase">
                <div className="bg-muted h-3 w-20 animate-pulse rounded" />
              </th>
              <th className="text-muted-foreground px-6 py-3 text-right text-xs font-medium tracking-wider uppercase">
                <div className="bg-muted ml-auto h-3 w-16 animate-pulse rounded" />
              </th>
            </tr>
          </thead>
          <tbody className="divide-y">
            {Array.from({ length: 4 }).map((_, i) => (
              <tr key={i}>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center gap-2">
                    <div className="bg-muted h-4 w-4 animate-pulse rounded" />
                    <div className="bg-muted h-4 w-32 animate-pulse rounded" />
                  </div>
                </td>
                <td className="px-6 py-4 text-right whitespace-nowrap">
                  <div className="bg-muted ml-auto h-8 w-16 animate-pulse rounded" />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
