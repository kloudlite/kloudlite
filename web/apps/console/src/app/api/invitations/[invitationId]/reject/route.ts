import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { rejectInvitation } from '@/lib/console/storage'

/**
 * POST /api/invitations/[invitationId]/reject
 * Reject an invitation
 */
export async function POST(
  _request: Request,
  { params }: { params: Promise<{ invitationId: string }> }
) {
  const { invitationId } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  try {
    await rejectInvitation(invitationId)
    return NextResponse.json({ success: true })
  } catch (_error) {
    return apiError('Failed to reject invitation', 500)
  }
}
