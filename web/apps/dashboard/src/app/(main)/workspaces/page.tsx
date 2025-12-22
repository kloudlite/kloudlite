import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkspacesList } from './_components/workspaces-list'
import { workspaceService } from '@/lib/services/workspace.service'
import { workMachineService } from '@/lib/services/work-machine.service'
import { getMyPreferences } from '@/app/actions/user-preferences.actions'
import type { Workspace } from '@kloudlite/types'

export default async function WorkspacesPage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Use username for filtering (matches ownedBy field in backend)
  const currentUser = session.user?.username || session.user?.email || 'user@example.com'

  // Fetch all data in parallel
  const [workMachineResult, workspacesResult, prefsResult] = await Promise.all([
    workMachineService.getMyWorkMachine().catch((err) => {
      console.error('Failed to fetch work machine:', err)
      return null
    }),
    workspaceService.listAll().catch((err) => {
      console.error('Failed to fetch workspaces:', err)
      return { items: [] }
    }),
    getMyPreferences(),
  ])

  // Extract work machine info
  const namespace = workMachineResult?.spec?.targetNamespace || 'default'
  const workMachineRunning = workMachineResult?.status?.state === 'running'

  // Extract workspaces
  const workspaces: Workspace[] = workspacesResult.items || []

  // Extract pinned workspace IDs from preferences
  const pinnedWorkspaceIds = new Set<string>()
  if (prefsResult.success && prefsResult.data?.spec.pinnedWorkspaces) {
    for (const ref of prefsResult.data.spec.pinnedWorkspaces) {
      pinnedWorkspaceIds.add(`${ref.namespace}/${ref.name}`)
    }
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
          namespace={namespace}
          workMachineRunning={workMachineRunning}
          pinnedWorkspaceIds={Array.from(pinnedWorkspaceIds)}
        />
      </div>
    </main>
  )
}
