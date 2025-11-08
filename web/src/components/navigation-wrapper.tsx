import { auth } from '@/lib/auth'
import { Navigation } from './navigation'
import { workMachineService } from '@/lib/services/work-machine.service'

export async function NavigationWrapper() {
  const session = await auth()
  const userRoles = session?.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const isAdmin = userRoles.includes('admin') || userRoles.includes('super-admin')

  // Check if user has a work machine
  let hasWorkMachine = false
  try {
    const workMachine = await workMachineService.getMyWorkMachine()
    hasWorkMachine = !!workMachine
  } catch (err) {
    // Silently handle the case where user doesn't have a work machine
    // This is expected for new users
    hasWorkMachine = false
  }

  return (
    <Navigation
      email={session?.user?.email}
      displayName={session?.user?.name}
      isSuperAdmin={isSuperAdmin}
      isAdmin={isAdmin}
      userRoles={userRoles}
      hasWorkMachine={hasWorkMachine}
    />
  )
}
