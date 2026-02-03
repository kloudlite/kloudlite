import { redirect } from 'next/navigation'
import Link from 'next/link'
import { getSession } from '@/lib/get-session'
import { EnvironmentsList } from './_components/environments-list'
import { listEnvironments } from '@/app/actions/environment.actions'
import { getMyWorkMachine } from '@/app/actions/work-machine.actions'
import { getMyPreferences } from '@/app/actions/user-preferences.actions'
import { environmentToUIModel, type EnvironmentUIModel } from '@kloudlite/types'
import { AlertCircle, Power } from 'lucide-react'
import { Alert, AlertDescription } from '@kloudlite/ui'
import { Button } from '@kloudlite/ui'

export default async function EnvironmentsPage() {
  const session = await getSession()

  if (!session) {
    redirect('/auth/signin')
  }

  // Use username for filtering (matches ownedBy field in backend)
  const currentUser = session.user?.username || session.user?.email || 'test-user'

  // Fetch data using server actions
  const [environmentsResult, workMachineResult, preferencesResult] = await Promise.all([
    listEnvironments(),
    getMyWorkMachine(),
    getMyPreferences(),
  ])

  const environments = environmentsResult.success ? environmentsResult.data : []
  const workMachine = workMachineResult.success ? workMachineResult.data : null
  const preferences = preferencesResult.success ? preferencesResult.data : null

  // Check if work machine is running
  const workMachineRunning = workMachine
    ? (workMachine.status?.state || workMachine.spec.state) === 'running' &&
      (workMachine.status?.isReady ?? false)
    : false

  // Get pinned environment IDs from preferences
  const pinnedEnvironmentIds = preferences?.pinnedEnvironments || []

  const allEnvironments: EnvironmentUIModel[] = environments.map((env) => {
    const owner = env.spec?.ownedBy || 'unknown'
    return environmentToUIModel(env, owner)
  })

  return (
    <>
      {/* Page Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-semibold tracking-tight mb-2">Environments</h1>
        <p className="text-muted-foreground text-sm">
          Manage development environments across your team
        </p>
      </div>

      {/* WorkMachine Status Banner */}
      {!workMachineRunning && (
        <Alert className="mb-6 border-warning bg-warning/10">
          <AlertCircle className="h-4 w-4 text-warning" />
          <AlertDescription className="flex items-center justify-between gap-4">
            <div className="flex-1">
              <p className="font-medium text-sm mb-1">WorkMachine is currently stopped</p>
              <p className="text-xs text-muted-foreground">
                Your development environments require a running WorkMachine to be accessible.
                Start your WorkMachine to connect to and manage your environments.
              </p>
            </div>
            <Link href="/">
              <Button size="sm" className="flex-shrink-0">
                <Power className="mr-2 h-4 w-4" />
                Start WorkMachine
              </Button>
            </Link>
          </AlertDescription>
        </Alert>
      )}

      {/* Environments List with Filter */}
      <EnvironmentsList
        environments={allEnvironments}
        currentUser={currentUser}
        workMachineRunning={workMachineRunning}
        pinnedEnvironmentIds={pinnedEnvironmentIds}
      />
    </>
  )
}
