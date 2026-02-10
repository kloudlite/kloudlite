import { UserManagementList } from '../_components/user-management-list'
import { listUsers } from '@/app/actions/user.actions'
import { listMachineTypes } from '@/app/actions/machine-type.actions'
import { listAllWorkMachines } from '@/app/actions/work-machine.actions'
import { UserDisplay, userToDisplay } from '@/types/user'
import { AlertCircle } from 'lucide-react'
import { getSession } from '@/lib/get-session'
import { redirect } from 'next/navigation'
import { env } from '@/lib/env'

// Error component
function UsersError({ error }: { error: string }) {
  return (
    <main className="space-y-6">
      <div>
        <h1 className="text-foreground text-2xl font-semibold">User Management</h1>
        <p className="mt-1.5 text-muted-foreground text-sm">Manage user accounts, roles, and permissions</p>
      </div>
      <div className="flex h-64 items-center justify-center">
        <div className="max-w-md text-center">
          <AlertCircle className="mx-auto mb-4 h-12 w-12 text-destructive" />
          <h3 className="mb-2 text-foreground text-lg font-medium">Failed to load users</h3>
          <p className="text-muted-foreground text-sm">{error}</p>
        </div>
      </div>
    </main>
  )
}

export default async function UsersPage() {
  // Check authentication and permissions
  const session = await getSession()
  if (!session || !session.user?.email) {
    redirect('/auth/signin')
  }

  // Check if user has admin or super-admin role
  const userRoles = session.user?.roles || []
  const isAdmin = userRoles.includes('admin')
  const isSuperAdmin = userRoles.includes('super-admin')

  if (!isAdmin && !isSuperAdmin) {
    redirect('/')
  }

  try {
    const isCloud = env.isKloudliteCloud

    // Fetch users, and optionally machine types + work machines for cloud mode
    const [usersResult, machineTypesResult, workMachinesResult] = await Promise.all([
      listUsers(),
      isCloud ? listMachineTypes() : Promise.resolve({ success: true, data: [] }),
      isCloud ? listAllWorkMachines() : Promise.resolve({ success: true, data: [] }),
    ])

    if (!usersResult.success) {
      return <UsersError error={usersResult.error || 'Unknown error occurred'} />
    }

    const users: UserDisplay[] = (usersResult.data || []).map(userToDisplay)

    // Transform machine types for the component (include tier details from annotations)
    const machineTypes = isCloud && machineTypesResult.success && machineTypesResult.data
      ? machineTypesResult.data
          .filter((mt: any) => mt.spec.active !== false)
          .map((mt: any) => {
            const ann = mt.metadata.annotations || {}
            return {
              id: mt.metadata.name,
              name: mt.spec.displayName || mt.metadata.name,
              description: mt.spec.description || '',
              cpu: mt.spec.resources?.cpu || '',
              memory: mt.spec.resources?.memory || '',
              tierSubtitle: ann['kloudlite.io/tier-subtitle'] || '',
              tierPrice: ann['kloudlite.io/tier-price'] || '',
              tierPriceUnit: ann['kloudlite.io/tier-price-unit'] || '',
              tierIncludedHours: ann['kloudlite.io/tier-included-hours'] || '',
              tierExtraHourPrice: ann['kloudlite.io/tier-extra-hour-price'] || '',
              tierStorageGb: ann['kloudlite.io/tier-storage-gb'] || '',
              tierSuspendMinutes: ann['kloudlite.io/tier-suspend-minutes'] || '',
              tierPopular: ann['kloudlite.io/tier-popular'] === 'true',
            }
          })
      : []

    // Build username → machineType map from work machines
    const workMachines = isCloud && workMachinesResult.success && workMachinesResult.data
      ? workMachinesResult.data
      : []

    return (
      <div className="mx-auto max-w-7xl space-y-6 px-6 py-8">
        {/* Page Header */}
        <div>
          <h1 className="text-foreground text-2xl font-semibold">User Management</h1>
          <p className="mt-1.5 text-muted-foreground text-sm">
            Manage user accounts, roles, and permissions
          </p>
        </div>

        {/* Users List Component */}
        <UserManagementList
          users={users}
          currentUserRole={isSuperAdmin ? 'super-admin' : 'admin'}
          isKloudliteCloud={isCloud}
          machineTypes={machineTypes}
          workMachines={workMachines}
        />
      </div>
    )
  } catch (_error) {
    return <UsersError error="An unexpected error occurred while loading users" />
  }
}
