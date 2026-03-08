export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { rejectOrgInvitation } from '@/lib/console/storage'

/**
 * POST /api/invitations/[invitationId]/reject
 * Reject an organization invitation
 */
export async function POST(
  _request: Request,
  { params }: { params: Promise<{ invitationId: string }> },
) {
  const { invitationId } = await params
  const session = await getRegistrationSession()

  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  try {
    await rejectOrgInvitation(invitationId, session.user.email)
    return NextResponse.json({ success: true })
  } catch (error) {
    return apiCatchError(error, 'Failed to reject invitation')
  }
}
