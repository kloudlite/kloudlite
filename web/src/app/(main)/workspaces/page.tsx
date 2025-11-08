import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkspacesList } from './_components/workspaces-list'
import { workspaceService } from '@/lib/services/workspace.service'
import { workMachineService } from '@/lib/services/work-machine.service'
import type { Workspace } from '@/types/workspace'

export default async function WorkspacesPage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  const currentUser = session.user?.email || 'user@example.com'

  // For demo, assume admin if email ends with @kloudlite.io
  const isAdmin = currentUser.endsWith('@kloudlite.io')

  // Get user's work machine to determine target namespace
  let namespace = 'default'

  try {
    const workMachine = await workMachineService.getMyWorkMachine()
    if (workMachine?.spec?.targetNamespace) {
      namespace = workMachine.spec.targetNamespace
    }
  } catch (err) {
    console.error('Failed to fetch work machine:', err)
    // Use default namespace if work machine is not found
    // This is expected for users who haven't set up a work machine yet
  }

  // Fetch real workspace data from API using the work machine's target namespace
  let workspaces: Workspace[] = []

  try {
    const response = await workspaceService.list(namespace)
    workspaces = response.items || []
  } catch (err) {
    console.error('Failed to fetch workspaces:', err)
    // Use empty array on error - user will see empty state
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
