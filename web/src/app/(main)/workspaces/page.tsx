import { redirect } from 'next/navigation'
import { auth } from '@/lib/auth'
import { WorkspacesList } from './_components/workspaces-list'
import { workspaceService } from '@/lib/services/workspace.service'
import { workMachineService } from '@/lib/services/work-machine.service'
import type { Workspace } from '@/types/workspace'

export default async function WorkspacesPage() {
  const session = await auth()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'user@example.com'

  // For demo, assume admin if email ends with @kloudlite.io
  const isAdmin = currentUser.endsWith('@kloudlite.io')

  // Get user's work machine to determine target namespace
  let namespace = 'default'
  let workMachineError = null

  try {
    const workMachine = await workMachineService.getMyWorkMachine()
    if (workMachine?.spec?.targetNamespace) {
      namespace = workMachine.spec.targetNamespace
    }
  } catch (err) {
    console.error('Failed to fetch work machine:', err)
    workMachineError = err instanceof Error ? err.message : 'Failed to fetch work machine'
  }

  // Fetch real workspace data from API using the work machine's target namespace
  let workspaces: Workspace[] = []
  let error: string | null = null

  try {
    const response = await workspaceService.list(namespace)
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
          <h1 className="text-2xl font-semibold">Workspaces</h1>
          <p className="text-muted-foreground mt-1.5 text-sm">
            Manage your development workspaces and collaborate with your team
          </p>
        </div>

        {/* Error Display */}
        {(error || workMachineError) && (
          <div className="mb-6 rounded-md border border-red-200 bg-red-50 p-4 dark:border-red-800 dark:bg-red-900/20">
            <div className="flex">
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800 dark:text-red-200">
                  {workMachineError ? 'Failed to load work machine' : 'Failed to load workspaces'}
                </h3>
                <div className="mt-2 text-sm text-red-700 dark:text-red-300">
                  <p>{workMachineError || error}</p>
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
          namespace={namespace}
        />
      </div>
    </main>
  )
}
