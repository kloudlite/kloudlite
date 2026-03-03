import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getErrorMessage } from '@/lib/errors'
import { getRegistrationSession } from '@/lib/console-auth'
import { acceptInvitation } from '@/lib/console/storage'

/**
 * POST /api/invitations/[invitationId]/accept
 * Accept an invitation
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
    const member = await acceptInvitation(invitationId, session.user.id)
    return NextResponse.json({ member })
  } catch (error) {
    return apiError(getErrorMessage(error, 'Failed to accept invitation'), 400)
  }
}
