import { Suspense } from 'react'
import { auth } from '@/lib/auth'
import { Navigation } from './navigation'
import { getMyWorkMachine } from '@/app/actions/work-machine.actions'
import { setThemeCookie } from '@/app/actions/theme'

/**
 * Inner navigation that receives work machine status
 */
async function NavigationWithWorkMachineStatus({
  email,
  displayName,
  isSuperAdmin,
  isAdmin,
  userRoles,
}: {
  email?: string
  displayName?: string
  isSuperAdmin: boolean
  isAdmin: boolean
  userRoles: string[]
}) {
  const workMachineResult = await getMyWorkMachine()
  const hasWorkMachine = workMachineResult.success && !!workMachineResult.data

  let isWorkMachineRunning = false
  if (hasWorkMachine && workMachineResult.data) {
    const state = workMachineResult.data.status?.state || workMachineResult.data.spec.state
    const isReady = workMachineResult.data.status?.isReady ?? false
    isWorkMachineRunning = state === 'running' && isReady
  }

  // Don't show navigation if user doesn't have a work machine
  if (!hasWorkMachine) {
    return null
  }

  return (
    <Navigation
      email={email}
      displayName={displayName}
      isSuperAdmin={isSuperAdmin}
      isAdmin={isAdmin}
      userRoles={userRoles}
      hasWorkMachine={hasWorkMachine}
      isWorkMachineRunning={isWorkMachineRunning}
      setThemeCookie={setThemeCookie}
    />
  )
}

/**
 * Skeleton for navigation while loading
 */
function NavigationSkeleton({
  email,
  displayName,
  isSuperAdmin,
  isAdmin,
  userRoles,
}: {
  email?: string
  displayName?: string
  isSuperAdmin: boolean
  isAdmin: boolean
  userRoles: string[]
}) {
  return (
    <Navigation
      email={email}
      displayName={displayName}
      isSuperAdmin={isSuperAdmin}
      isAdmin={isAdmin}
      userRoles={userRoles}
      hasWorkMachine={true} // Assume true for skeleton to show full nav
      isWorkMachineRunning={false}
      setThemeCookie={setThemeCookie}
    />
  )
}

export async function NavigationWrapper() {
  // Session data loads quickly - this is the critical path for user profile
  const session = await auth()
  const userRoles = session?.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const isAdmin = userRoles.includes('admin') || userRoles.includes('super-admin')
  const email = session?.user?.email
  const displayName = session?.user?.name

  return (
    <Suspense
      fallback={
        <NavigationSkeleton
          email={email}
          displayName={displayName}
          isSuperAdmin={isSuperAdmin}
          isAdmin={isAdmin}
          userRoles={userRoles}
        />
      }
    >
      <NavigationWithWorkMachineStatus
        email={email}
        displayName={displayName}
        isSuperAdmin={isSuperAdmin}
        isAdmin={isAdmin}
        userRoles={userRoles}
      />
    </Suspense>
  )
}
