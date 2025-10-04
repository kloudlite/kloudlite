import { auth } from '@/lib/auth'
import { Navigation } from './navigation'

export async function NavigationWrapper() {
  const session = await auth()
  const userRoles = session?.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const isAdmin = userRoles.includes('admin') || userRoles.includes('super-admin')

  return (
    <Navigation
      email={session?.user?.email}
      displayName={session?.user?.name}
      isSuperAdmin={isSuperAdmin}
      isAdmin={isAdmin}
      userRoles={userRoles}
    />
  )
}