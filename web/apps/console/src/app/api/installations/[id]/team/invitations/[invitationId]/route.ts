import { NextResponse } from 'next/server'
import { apiCatchError } from '@/lib/api-helpers'
import { requireManagePermission } from '@/lib/console/authorization'
import { deleteInvitation } from '@/lib/console/storage'

/**
 * DELETE /api/installations/[id]/team/invitations/[invitationId]
 * Cancel/delete invitation (owner/admin only)
 */
export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ id: string; invitationId: string }> }
) {
  const { id, invitationId } = await params

  try {
    await requireManagePermission(id)
    await deleteInvitation(invitationId)

    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to delete invitation')
  }
}
