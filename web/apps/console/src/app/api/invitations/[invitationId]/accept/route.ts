export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { acceptOrgInvitation } from '@/lib/console/storage'

/**
 * POST /api/invitations/[invitationId]/accept
 * Accept an organization invitation
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
    const member = await acceptOrgInvitation(invitationId, session.user.id, session.user.email)
    return NextResponse.json({ member })
  } catch (error) {
    return apiCatchError(error, 'Failed to accept invitation')
  }
}
