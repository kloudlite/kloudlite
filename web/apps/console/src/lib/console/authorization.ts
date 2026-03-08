/**
 * Authorization helpers for org-level role-based access control
 */

import { getRegistrationSession } from '@/lib/console-auth'
import { getOrgMemberRole, type OrgRole } from '@/lib/console/storage'
import { getInstallationById } from '@/lib/console/storage'

export interface AuthContext {
  userId: string
  orgId: string
  role: OrgRole
}

/**
 * Verify user is a member of the organization and return their role
 * Throws error if unauthorized
 */
export async function requireOrgAccess(orgId: string): Promise<AuthContext> {
  const session = await getRegistrationSession()

  if (!session?.user) {
    throw new Error('Unauthorized: No session')
  }

  const role = await getOrgMemberRole(orgId, session.user.id)

  if (!role) {
    throw new Error('Forbidden: Not a member of this organization')
  }

  return {
    userId: session.user.id,
    orgId,
    role,
  }
}

/**
 * Require owner role for the organization
 */
export async function requireOrgOwner(orgId: string): Promise<AuthContext> {
  const context = await requireOrgAccess(orgId)

  if (context.role !== 'owner') {
    throw new Error('Forbidden: Requires owner role')
  }

  return context
}

/**
 * Verify user has access to an installation via org membership
 * Looks up the installation's org_id and checks org membership
 */
export async function requireInstallationAccess(
  installationId: string,
): Promise<AuthContext & { installationId: string }> {
  const session = await getRegistrationSession()

  if (!session?.user) {
    throw new Error('Unauthorized: No session')
  }

  const installation = await getInstallationById(installationId)
  if (!installation) {
    throw new Error('Not found: Installation does not exist')
  }

  const role = await getOrgMemberRole(installation.orgId, session.user.id)
  if (!role) {
    throw new Error('Forbidden: Not a member of the organization that owns this installation')
  }

  return {
    userId: session.user.id,
    orgId: installation.orgId,
    role,
    installationId,
  }
}

/**
 * Require owner role for the org that owns the installation
 */
export async function requireInstallationOwner(
  installationId: string,
): Promise<AuthContext & { installationId: string }> {
  const context = await requireInstallationAccess(installationId)

  if (context.role !== 'owner') {
    throw new Error('Forbidden: Requires owner role')
  }

  return context
}
