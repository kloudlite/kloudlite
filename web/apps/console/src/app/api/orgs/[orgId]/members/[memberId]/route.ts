export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiCatchError, apiError } from '@/lib/api-helpers'
import { requireOrgAccess } from '@/lib/console/authorization'
import { getOrgMembers, removeOrgMember } from '@/lib/console/storage'

/**
 * DELETE /api/orgs/[orgId]/members/[memberId]
 * Remove a member from the organization
 * Owners cannot be removed. Only owners and admins can remove other admins.
 */
export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ orgId: string; memberId: string }> },
) {
  const { orgId, memberId } = await params

  try {
    await requireOrgAccess(orgId)

    // Validate member belongs to this org
    const members = await getOrgMembers(orgId)
    const member = members.find(m => m.id === memberId)
    if (!member) return apiError('Member not found in this organization', 404)
    if (member.role === 'owner') return apiError('Cannot remove the organization owner', 403)

    await removeOrgMember(memberId)
    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to remove member')
  }
}
