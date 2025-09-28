import { auth } from '@/lib/auth'
import { Navigation } from './navigation'

export async function NavigationWrapper() {
  const session = await auth()
  const userRoles = session?.user?.roles || []
  const isSuperAdmin = userRoles.includes('super-admin')
  const platformRole = session?.user?.platformRole
  const isAdmin = platformRole === 'admin' || platformRole === 'super_admin'

  return (
    <Navigation
      email={session?.user?.email}
      isSuperAdmin={isSuperAdmin}
      isAdmin={isAdmin}
    />
  )
}