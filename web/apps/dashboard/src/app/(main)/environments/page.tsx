import { redirect } from 'next/navigation'
import { getSession } from '@/lib/get-session'
import { EnvironmentsList } from './_components/environments-list'
import { getEnvironmentsListFull } from '@/lib/services/dashboard.service'
import { environmentToUIModel, type EnvironmentUIModel } from '@kloudlite/types'

export default async function EnvironmentsPage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Use username for filtering (matches ownedBy field in backend)
  const currentUser = session.user?.username || session.user?.email || 'test-user'

  // Single API call to get environments, work machine, and preferences
  const data = await getEnvironmentsListFull().catch((err) => {
    console.error('Failed to fetch environments list:', err)
    return {
      environments: [],
      workMachine: null,
      preferences: null,
      pinnedEnvironmentIds: [],
      workMachineRunning: false,
    }
  })

  const allEnvironments: EnvironmentUIModel[] = (data.environments || []).map((env) => {
    const owner = env.spec?.ownedBy || 'unknown'
    return environmentToUIModel(env, owner)
  })

  return (
    <main className="mx-auto max-w-7xl px-6 py-8">
      {/* Title and Filter Section */}
      <div className="mb-8">
        <div className="mb-6">
          <h1 className="text-2xl font-semibold">Environments</h1>
          <p className="text-muted-foreground mt-1.5 text-sm">
            Manage development environments across your team
          </p>
        </div>

        {/* Environments List with Filter */}
        <EnvironmentsList
          environments={allEnvironments}
          currentUser={currentUser}
          workMachineRunning={data.workMachineRunning}
          pinnedEnvironmentIds={data.pinnedEnvironmentIds}
        />
      </div>
    </main>
  )
}
