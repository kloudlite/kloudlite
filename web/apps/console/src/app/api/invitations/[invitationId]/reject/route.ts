import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { rejectInvitation } from '@/lib/console/supabase-storage-service'

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
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  try {
    await rejectInvitation(invitationId)
    return NextResponse.json({ success: true })
  } catch (error) {
    return NextResponse.json({ error: 'Failed to reject invitation' }, { status: 500 })
  }
}
