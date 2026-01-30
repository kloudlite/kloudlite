import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getUserPendingInvitations } from '@/lib/console/storage'

/**
 * GET /api/invitations/my
 * Get user's pending invitations
 */
export async function GET() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  try {
    const invitations = await getUserPendingInvitations(session.user.email)
    return NextResponse.json({ invitations })
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to get invitations' },
      { status: 500 }
    )
  }
}
