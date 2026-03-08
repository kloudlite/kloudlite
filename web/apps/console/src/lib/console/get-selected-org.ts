import { cookies } from 'next/headers'
import { getUserOrganizations, createOrganization } from '@/lib/console/storage'
import type { Organization } from '@/lib/console/storage'

const COOKIE_NAME = 'selected_org_id'

/**
 * Get the currently selected organization for the authenticated user.
 * Reads from the `selected_org_id` cookie, validates the user is still a member,
 * and falls back to the first org if cookie is stale/missing.
 *
 * If the user has no orgs, attempts to auto-create one.
 * Returns null only if org creation also fails.
 */
export async function getSelectedOrg(
  userId: string,
  userName?: string,
  userEmail?: string,
): Promise<Organization | null> {
  let orgs = await getUserOrganizations(userId)

  // Auto-create org for users with none (pre-migration users)
  if (orgs.length === 0 && userEmail) {
    try {
      const baseSlug = (userName || userEmail.split('@')[0])
        .toLowerCase()
        .replace(/[^a-z0-9-]/g, '-')
        .replace(/-+/g, '-')
        .replace(/^-|-$/g, '')
        .slice(0, 50)
      let slug = /^[a-z]/.test(baseSlug) ? baseSlug : `org-${baseSlug}`
      if (slug.length < 3) slug = `${slug}-org`

      await createOrganization(userId, `${userName || 'My'}'s Organization`, slug)
      orgs = await getUserOrganizations(userId)
    } catch {
      // Best-effort
    }
  }

  if (orgs.length === 0) return null

  // Check cookie for selected org
  const cookieStore = await cookies()
  const selectedOrgId = cookieStore.get(COOKIE_NAME)?.value

  if (selectedOrgId) {
    const selectedOrg = orgs.find((o) => o.id === selectedOrgId)
    if (selectedOrg) return selectedOrg
  }

  // Fall back to first org
  return orgs[0]
}
