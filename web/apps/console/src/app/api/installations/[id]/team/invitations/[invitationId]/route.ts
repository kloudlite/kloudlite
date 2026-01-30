import { NextResponse } from 'next/server'
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
    const message = error instanceof Error ? error.message : 'Failed to delete invitation'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
