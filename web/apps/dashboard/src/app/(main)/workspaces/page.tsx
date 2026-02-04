import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkspacesList } from './_components/workspaces-list'
import { getWorkspacesListFull } from '@/app/actions/workspace.actions'
import { WorkMachineStoppedAlert } from '@/components/work-machine-stopped-alert'
import type { Workspace } from '@kloudlite/types'

export default async function WorkspacesPage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Use username for filtering (matches ownedBy field in backend)
  const currentUser = session.user?.username || session.user?.email || 'user@example.com'

  // Fetch workspaces, work machine, and preferences using server action
  const result = await getWorkspacesListFull()
  const data = result.data // Always has data, even on error (fallback values)

  // Extract work machine namespace
  const namespace = data.workMachine?.spec?.targetNamespace || 'default'

  return (
    <>
      {/* Page Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight mb-2">Workspaces</h1>
        <p className="text-muted-foreground text-sm">
          Manage your development workspaces and collaborate with your team
        </p>
      </div>

      {/* WorkMachine Status Banner */}
      {!data.workMachineRunning && <WorkMachineStoppedAlert />}

      {/* Workspaces List with Filter */}
      <WorkspacesList
        workspaces={(data.workspaces || []) as Workspace[]}
        currentUser={currentUser}
        namespace={namespace}
        workMachineRunning={data.workMachineRunning}
        pinnedWorkspaceIds={data.pinnedWorkspaceIds || []}
      />
    </>
  )
}
