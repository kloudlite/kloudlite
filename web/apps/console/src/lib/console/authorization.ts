/**
 * Authorization helpers for role-based access control
 */

import { getRegistrationSession } from '@/lib/console-auth'
import { getMemberRole, type MemberRole } from '@/lib/console/storage'

export interface AuthContext {
  userId: string
  installationId: string
  role: MemberRole
}

/**
 * Verify user has access to installation and return their role
 * Throws error if unauthorized
 */
export async function requireInstallationAccess(
  installationId: string
): Promise<AuthContext> {
  const session = await getRegistrationSession()

  if (!session?.user) {
    throw new Error('Unauthorized: No session')
  }

  const role = await getMemberRole(installationId, session.user.id)

  if (!role) {
    throw new Error('Forbidden: Not a member of this installation')
  }

  return {
    userId: session.user.id,
    installationId,
    role,
  }
}

/**
 * Require owner or admin role
 */
export async function requireManagePermission(
  installationId: string
): Promise<AuthContext> {
  const context = await requireInstallationAccess(installationId)

  if (context.role !== 'owner' && context.role !== 'admin') {
    throw new Error('Forbidden: Requires owner or admin role')
  }

  return context
}

/**
 * Require owner role only
 */
export async function requireOwnerPermission(
  installationId: string
): Promise<AuthContext> {
  const context = await requireInstallationAccess(installationId)

  if (context.role !== 'owner') {
    throw new Error('Forbidden: Requires owner role')
  }

  return context
}

/**
 * Permission helper for role hierarchy
 */
export function hasPermission(userRole: MemberRole, requiredRole: MemberRole): boolean {
  const hierarchy: Record<MemberRole, number> = {
    owner: 4,
    admin: 3,
    member: 2,
    viewer: 1,
  }

  return hierarchy[userRole] >= hierarchy[requiredRole]
}
