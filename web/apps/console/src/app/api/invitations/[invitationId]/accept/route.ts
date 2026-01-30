import { NextResponse } from 'next/server'
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
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  try {
    const member = await acceptInvitation(invitationId, session.user.id)
    return NextResponse.json({ member })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to accept invitation'
    return NextResponse.json({ error: message }, { status: 400 })
  }
}
