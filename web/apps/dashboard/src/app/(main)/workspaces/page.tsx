import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { WorkspacesList } from './_components/workspaces-list'
import { getWorkspacesListFull } from '@/lib/services/dashboard.service'

export default async function WorkspacesPage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Use username for filtering (matches ownedBy field in backend)
  const currentUser = session.user?.username || session.user?.email || 'user@example.com'

  // Single API call to get workspaces, work machine, and preferences
  const data = await getWorkspacesListFull().catch((err) => {
    console.error('Failed to fetch workspaces list:', err)
    return {
      workspaces: [],
      workMachine: null,
      preferences: null,
      pinnedWorkspaceIds: [],
      workMachineRunning: false,
    }
  })

  // Extract work machine namespace
  const namespace = data.workMachine?.spec?.targetNamespace || 'default'

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
          workspaces={data.workspaces || []}
          currentUser={currentUser}
          namespace={namespace}
          workMachineRunning={data.workMachineRunning}
          pinnedWorkspaceIds={data.pinnedWorkspaceIds || []}
        />
      </div>
    </main>
  )
}
