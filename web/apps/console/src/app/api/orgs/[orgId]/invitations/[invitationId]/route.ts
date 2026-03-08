export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiCatchError, apiError } from '@/lib/api-helpers'
import { requireOrgAccess } from '@/lib/console/authorization'
import { deleteOrgInvitation, getOrgInvitations } from '@/lib/console/storage'

/**
 * DELETE /api/orgs/[orgId]/invitations/[invitationId]
 * Cancel a pending invitation
 */
export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ orgId: string; invitationId: string }> },
) {
  const { orgId, invitationId } = await params

  try {
    await requireOrgAccess(orgId)

    // Validate invitation belongs to this org
    const invitations = await getOrgInvitations(orgId)
    const invitation = invitations.find(i => i.id === invitationId)
    if (!invitation) return apiError('Invitation not found in this organization', 404)

    await deleteOrgInvitation(invitationId)
    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to cancel invitation')
  }
}
