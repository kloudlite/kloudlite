import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { WorkspacesList } from '@/components/workspaces-list'
import { workspaceService } from '@/services/workspace-service'

export default async function WorkspacesPage() {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'user@example.com'

  // For demo, assume admin if email ends with @kloudlite.io
  const isAdmin = currentUser.endsWith('@kloudlite.io')

  // Fetch real workspace data from API
  let workspaces = []
  let error = null

  try {
    const response = await workspaceService.list('default')
    workspaces = response.items || []
  } catch (err) {
    console.error('Failed to fetch workspaces:', err)
    error = err instanceof Error ? err.message : 'Failed to fetch workspaces'
    // Empty workspaces array if API fails
    workspaces = []
  }

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Title and Filter Section */}
      <div className="mb-8">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold text-gray-900">Workspaces</h1>
          <p className="text-sm text-gray-600 mt-1.5">
            Manage your development workspaces and collaborate with your team
          </p>
        </div>

        {/* Error Display */}
        {error && (
          <div className="mb-6 rounded-md bg-red-50 border border-red-200 p-4">
            <div className="flex">
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  Failed to load workspaces
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{error}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Workspaces List with Filter */}
        <WorkspacesList
          workspaces={workspaces}
          currentUser={currentUser}
          isAdmin={isAdmin}
          namespace="default"
        />
      </div>
    </main>
  )
}