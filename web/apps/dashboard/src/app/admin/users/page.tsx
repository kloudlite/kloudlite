import { UserManagementList } from '../_components/user-management-list'
import { getAllUsers } from '@/lib/actions/user-actions'
import { UserDisplay } from '@/types/user'
import { AlertCircle } from 'lucide-react'
import { getSession } from '@/lib/get-session'
import { redirect } from 'next/navigation'

// Error component
function UsersError({ error }: { error: string }) {
  return (
    <main className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">User Management</h1>
        <p className="mt-1.5 text-sm text-gray-600">Manage user accounts, roles, and permissions</p>
      </div>
      <div className="flex h-64 items-center justify-center">
        <div className="max-w-md text-center">
          <AlertCircle className="mx-auto mb-4 h-12 w-12 text-red-500" />
          <h3 className="mb-2 text-lg font-medium text-gray-900">Failed to load users</h3>
          <p className="text-sm text-gray-600">{error}</p>
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
    const result = await getAllUsers()

    if (!result.success) {
      return <UsersError error={result.error || 'Unknown error occurred'} />
    }

    const users: UserDisplay[] = result.users || []

    return (
      <div className="mx-auto max-w-7xl space-y-6 px-6 py-8">
        {/* Page Header */}
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">User Management</h1>
          <p className="mt-1.5 text-sm text-gray-600">
            Manage user accounts, roles, and permissions
          </p>
        </div>

        {/* Users List Component */}
        <UserManagementList
          users={users}
          currentUserRole={isSuperAdmin ? 'super-admin' : 'admin'}
        />
      </div>
    )
  } catch (_error) {
    return <UsersError error="An unexpected error occurred while loading users" />
  }
}
