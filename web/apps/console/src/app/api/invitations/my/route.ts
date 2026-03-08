export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { getUserPendingOrgInvitations } from '@/lib/console/storage'

/**
 * GET /api/invitations/my
 * Get current user's pending organization invitations
 */
export async function GET() {
  const session = await getRegistrationSession()

  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  try {
    const invitations = await getUserPendingOrgInvitations(session.user.email)
    return NextResponse.json({ invitations })
  } catch (error) {
    return apiCatchError(error, 'Failed to get invitations')
  }
}
