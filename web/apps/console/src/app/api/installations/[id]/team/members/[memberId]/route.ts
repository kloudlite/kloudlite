import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireManagePermission } from '@/lib/console/authorization'
import {
  removeInstallationMember,
  updateMemberRole,
  type MemberRole,
} from '@/lib/console/storage'

/**
 * PATCH /api/installations/[id]/team/members/[memberId]
 * Update member role (owner/admin only)
 */
export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ id: string; memberId: string }> }
) {
  const { id, memberId } = await params

  try {
    await requireManagePermission(id)
    const body = await request.json()
    const { role } = body as { role: MemberRole }

    if (!role || !['owner', 'admin', 'member', 'viewer'].includes(role)) {
      return apiError('Invalid role', 400)
    }

    await updateMemberRole(memberId, role)

    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to update member')
  }
}

/**
 * DELETE /api/installations/[id]/team/members/[memberId]
 * Remove member (owner/admin only, cannot remove owner)
 */
export async function DELETE(
  _request: Request,
  { params }: { params: Promise<{ id: string; memberId: string }> }
) {
  const { id, memberId } = await params

  try {
    await requireManagePermission(id)
    await removeInstallationMember(memberId)

    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to remove member')
  }
}
