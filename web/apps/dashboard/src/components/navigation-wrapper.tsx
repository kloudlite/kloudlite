import { auth } from '@/lib/auth'
import { Navigation } from './navigation'
import { getMyWorkMachine } from '@/app/actions/work-machine.actions'
import { setThemeCookie } from '@/app/actions/theme'

export async function NavigationWrapper() {
  const session = await auth()
  const userRoles = session?.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const isAdmin = userRoles.includes('admin') || userRoles.includes('super-admin')

  // Check if user has a work machine and if it's running
  const workMachineResult = await getMyWorkMachine()
  const hasWorkMachine = workMachineResult.success && !!workMachineResult.data

  let isWorkMachineRunning = false
  if (hasWorkMachine && workMachineResult.data) {
    const state = workMachineResult.data.status?.state || workMachineResult.data.spec.state
    const isReady = workMachineResult.data.status?.isReady ?? false
    isWorkMachineRunning = state === 'running' && isReady
  }

  // Don't show navigation if user doesn't have a work machine
  // They need to complete initial setup first
  if (!hasWorkMachine) {
    return null
  }

  return (
    <Navigation
      email={session?.user?.email}
      displayName={session?.user?.name}
      isSuperAdmin={isSuperAdmin}
      isAdmin={isAdmin}
      userRoles={userRoles}
      hasWorkMachine={hasWorkMachine}
      isWorkMachineRunning={isWorkMachineRunning}
      setThemeCookie={setThemeCookie}
    />
  )
}
