// Frontend display user interface for the user management table
export interface UserDisplay {
  id: string
  name: string
  email: string
  role: string
  status: string
  lastLogin: string
  created: string
  firstName?: string
  lastName?: string
  displayName?: string
  providers?: Array<{
    provider: string
    providerId: string
    connectedAt: string
  }>
}

// Form data interfaces
export interface CreateUserFormData {
  email: string
  displayName?: string
  roles: string[]
}

export interface UpdateUserFormData {
  email?: string
  displayName?: string
  firstName?: string
  lastName?: string
  isActive?: boolean
}

// Utility function to convert API User to UserDisplay
export function userToDisplay(user: any): UserDisplay {
  const email = user.spec?.email || ''
  const displayName = user.spec?.displayName ||
    (user.spec?.firstName && user.spec?.lastName
      ? `${user.spec.firstName} ${user.spec.lastName}`.trim()
      : user.spec?.firstName || user.spec?.lastName || email.split('@')[0])

  const createdAt = user.metadata?.creationTimestamp
    ? new Date(user.metadata.creationTimestamp)
    : new Date()

  const lastLoginAt = user.status?.lastLogin
    ? new Date(user.status.lastLogin)
    : null

  // Extract roles from the roles array or default to ['user']
  const roles = user.spec?.roles || []

  // For display, show the highest privilege role first, but store all roles
  const roleHierarchy = ['super-admin', 'admin', 'user']
  const userRoles = roles.length > 0 ? roles : ['user']
  const primaryRole = roleHierarchy.find(role => userRoles.includes(role)) || userRoles[0]

  return {
    id: user.metadata?.name || user.metadata?.uid || email,
    name: displayName,
    email: email,
    role: userRoles.length > 1 ? userRoles.join(', ') : primaryRole, // Show all roles if multiple
    status: user.spec?.active !== false ? 'active' : 'inactive',
    lastLogin: lastLoginAt ? formatTimeAgo(lastLoginAt) : 'Never',
    created: formatTimeAgo(createdAt),
    firstName: user.spec?.firstName,
    lastName: user.spec?.lastName,
    displayName: user.spec?.displayName,
    providers: user.spec?.providers?.map((p: any) => ({
      provider: p.provider,
      providerId: p.providerId,
      connectedAt: p.connectedAt
    })) || []
  }
}

// Helper function to format time ago
function formatTimeAgo(date: Date): string {
  const now = new Date()
  const diffInMs = now.getTime() - date.getTime()
  const diffInMinutes = Math.floor(diffInMs / (1000 * 60))
  const diffInHours = Math.floor(diffInMinutes / 60)
  const diffInDays = Math.floor(diffInHours / 24)
  const diffInMonths = Math.floor(diffInDays / 30)

  if (diffInMinutes < 1) return 'Just now'
  if (diffInMinutes < 60) return `${diffInMinutes} mins ago`
  if (diffInHours < 24) return `${diffInHours} hours ago`
  if (diffInDays < 30) return `${diffInDays} days ago`
  return `${diffInMonths} months ago`
}