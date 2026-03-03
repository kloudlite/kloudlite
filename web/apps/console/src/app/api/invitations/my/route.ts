import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { getUserPendingInvitations } from '@/lib/console/storage'

/**
 * GET /api/invitations/my
 * Get user's pending invitations
 */
export async function GET() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  try {
    const invitations = await getUserPendingInvitations(session.user.email)
    return NextResponse.json({ invitations })
  } catch (error) {
    return apiError('Failed to get invitations', 500)
  }
}
