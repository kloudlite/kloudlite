import { Skeleton } from '@/components/ui/skeleton'
import { Button } from '@/components/ui/button'
import { Search, Plus } from 'lucide-react'

export default function TeamsLoading() {
  return (
    <div className="min-h-screen bg-muted/30 flex flex-col">
      {/* Header */}
      <div className="bg-background border-b">
        <div className="container mx-auto px-6 py-4 max-w-7xl">
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <Skeleton className="h-8 w-32" />
              <Skeleton className="h-4 w-48" />
            </div>
            <div className="flex items-center gap-3">
              <Button variant="ghost" size="sm" disabled>
                <Search className="h-4 w-4 mr-2" />
                Search Teams
              </Button>
              <Skeleton className="h-8 w-8 rounded-full" />
            </div>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1">
        <div className="container mx-auto px-6 py-6 max-w-7xl space-y-6">
          {/* Pending Invitations Skeleton */}
          <div>
            <Skeleton className="h-6 w-40 mb-4" />
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {[1, 2].map((i) => (
                <div key={i} className="bg-background border rounded-sm p-4">
                  <div className="flex items-start justify-between mb-3">
                    <div className="flex-1 space-y-2">
                      <Skeleton className="h-5 w-32" />
                      <Skeleton className="h-4 w-full" />
                    </div>
                    <Skeleton className="h-5 w-16 rounded-sm" />
                  </div>
                  <Skeleton className="h-3 w-40 mb-4" />
                  <div className="flex gap-2">
                    <Skeleton className="h-8 w-16" />
                    <Skeleton className="h-8 w-16" />
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* All Teams Section */}
          <div className="bg-background border rounded-sm">
            {/* Table Header */}
            <div className="border-b px-6 py-4">
              <div className="flex items-center justify-between">
                <Skeleton className="h-6 w-24" />
                <div className="flex items-center gap-2">
                  <Button variant="ghost" size="sm" disabled>
                    Sort
                  </Button>
                  <Button size="sm" disabled>
                    <Plus className="h-4 w-4 mr-2" />
                    New Team
                  </Button>
                </div>
              </div>
            </div>

            {/* Table Skeleton */}
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead>
                  <tr className="border-b text-sm text-muted-foreground">
                    <th className="text-left font-medium px-6 py-3">Team Name</th>
                    <th className="text-left font-medium px-6 py-3 hidden sm:table-cell">Role</th>
                    <th className="text-left font-medium px-6 py-3 hidden md:table-cell">Members</th>
                    <th className="text-left font-medium px-6 py-3 hidden lg:table-cell">Last Accessed</th>
                    <th className="text-left font-medium px-6 py-3 hidden lg:table-cell">Member Since</th>
                    <th className="px-6 py-3 w-16"></th>
                  </tr>
                </thead>
                <tbody className="divide-y">
                  {[1, 2, 3, 4, 5].map((i) => (
                    <tr key={i}>
                      <td className="px-6 py-4">
                        <div className="space-y-2">
                          <Skeleton className="h-5 w-32" />
                          <Skeleton className="h-4 w-48" />
                        </div>
                      </td>
                      <td className="px-6 py-4 hidden sm:table-cell">
                        <Skeleton className="h-4 w-16" />
                      </td>
                      <td className="px-6 py-4 hidden md:table-cell">
                        <Skeleton className="h-4 w-8" />
                      </td>
                      <td className="px-6 py-4 hidden lg:table-cell">
                        <Skeleton className="h-4 w-20" />
                      </td>
                      <td className="px-6 py-4 hidden lg:table-cell">
                        <Skeleton className="h-4 w-20" />
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center justify-end gap-2">
                          <Skeleton className="h-8 w-8" />
                          <Skeleton className="h-8 w-8" />
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      {/* Footer */}
      <footer className="border-t border-border mt-auto">
        <div className="max-w-6xl mx-auto px-6 py-12">
          <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-8">
            <div className="flex items-center gap-4">
              <Skeleton className="h-4 w-48" />
            </div>
            <div className="flex flex-wrap gap-6">
              {[1, 2, 3, 4].map((i) => (
                <Skeleton key={i} className="h-4 w-16" />
              ))}
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}