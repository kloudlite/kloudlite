import { Navigation } from './navigation'
import { setThemeCookie } from '@/app/actions/theme'

interface NavigationWrapperProps {
  email?: string
  displayName?: string
  isSuperAdmin: boolean
  isAdmin: boolean
  userRoles: string[]
  hasWorkMachine: boolean
  isWorkMachineRunning: boolean
}

export function NavigationWrapper({
  email,
  displayName,
  isSuperAdmin,
  isAdmin,
  userRoles,
  hasWorkMachine,
  isWorkMachineRunning,
}: NavigationWrapperProps) {
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
