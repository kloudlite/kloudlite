import { NextResponse } from 'next/server'
import {
  requireManagePermission,
  requireInstallationAccess,
} from '@/lib/console/authorization'
import {
  createInvitation,
  getInstallationInvitations,
  getUserByEmail,
  type MemberRole,
} from '@/lib/console/supabase-storage-service'

/**
 * GET /api/installations/[id]/team/invitations
 * List pending invitations
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params

  try {
    await requireInstallationAccess(id)
    const invitations = await getInstallationInvitations(id)

    return NextResponse.json({ invitations })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to get invitations'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}

/**
 * POST /api/installations/[id]/team/invitations
 * Create new invitation (owner/admin only)
 */
export async function POST(
  request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params

  try {
    const context = await requireManagePermission(id)
    const body = await request.json()
    const { email, role } = body as { email: string; role: Exclude<MemberRole, 'owner'> }

    if (!email || !role) {
      return NextResponse.json(
        { error: 'Email and role are required' },
        { status: 400 }
      )
    }

    if (role === 'owner') {
      return NextResponse.json(
        { error: 'Cannot invite as owner. Use transfer ownership instead.' },
        { status: 400 }
      )
    }

    // Check if user already exists and is a member
    const existingUser = await getUserByEmail(email)
    if (existingUser) {
      // The invitation will fail on constraint if user is already a member
      // This is handled by the database unique constraint
    }

    const invitation = await createInvitation(id, email, role, context.userId)

    return NextResponse.json({ invitation })
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Failed to create invitation'
    const status = message.includes('Unauthorized')
      ? 401
      : message.includes('Forbidden')
        ? 403
        : message.includes('already has')
          ? 409
          : 500
    return NextResponse.json({ error: message }, { status })
  }
}
