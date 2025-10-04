import { UserManagementList } from '../_components/user-management-list'
import { getAllUsers } from '@/lib/actions/user-actions'
import { UserDisplay } from '@/types/user'
import { AlertCircle } from 'lucide-react'
import { auth } from '@/lib/auth'
import { redirect } from 'next/navigation'

// Error component
function UsersError({ error }: { error: string }) {
  return (
    <main className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-gray-900">User Management</h1>
        <p className="text-sm text-gray-600 mt-1.5">
          Manage user accounts, roles, and permissions
        </p>
      </div>
      <div className="flex items-center justify-center h-64">
        <div className="text-center max-w-md">
          <AlertCircle className="h-12 w-12 text-red-500 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-gray-900 mb-2">Failed to load users</h3>
          <p className="text-sm text-gray-600">{error}</p>
        </div>
      </div>
    </main>
  )
}

export default async function UsersPage() {
  // Check authentication and permissions
  const session = await auth()
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
      <div className="mx-auto max-w-7xl px-6 py-8 space-y-6">
        {/* Page Header */}
        <div>
          <h1 className="text-2xl font-semibold text-gray-900">User Management</h1>
          <p className="text-sm text-gray-600 mt-1.5">
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
  } catch (error) {
    return <UsersError error="An unexpected error occurred while loading users" />
  }
}