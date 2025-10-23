// Frontend display user interface for the user management table
export interface UserDisplay {
  id: string
  name: string
  email: string
  role: string
  status: string
  lastLogin: string
  created: string
  displayName?: string
  providers?: Array<{
    provider: string
    providerId: string
    connectedAt?: string
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
  isActive?: boolean
}

// Backend API User resource structure
export interface UserProvider {
  provider: string
  providerId: string
  connectedAt?: string
}

export interface UserResource {
  metadata?: {
    name?: string
    uid?: string
    creationTimestamp?: string
  }
  spec?: {
    email?: string
    displayName?: string
    active?: boolean
    roles?: string[]
    providers?: UserProvider[]
  }
  status?: {
    lastLogin?: string
  }
}

// Utility function to convert API User to UserDisplay
export function userToDisplay(user: UserResource): UserDisplay {
  const email = user.spec?.email || ''
  const displayName = user.spec?.displayName || email.split('@')[0]

  const createdAt = user.metadata?.creationTimestamp
    ? new Date(user.metadata.creationTimestamp)
    : new Date()

  const lastLoginAt = user.status?.lastLogin
    ? new Date(user.status.lastLogin)
    : null

  // Extract roles from the roles array or default to ['user']
  const roles = user.spec?.roles || []

  // For display, show all roles consistently
  const userRoles = roles.length > 0 ? roles : ['user']

  return {
    id: user.metadata?.name || user.metadata?.uid || email,
    name: displayName,
    email: email,
    role: userRoles.join(', '), // Always show all roles consistently
    status: user.spec?.active !== false ? 'active' : 'inactive',
    lastLogin: lastLoginAt ? formatTimeAgo(lastLoginAt) : 'Never',
    created: formatTimeAgo(createdAt),
    displayName: user.spec?.displayName,
    providers: user.spec?.providers?.map((p: UserProvider) => ({
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