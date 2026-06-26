export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireOrgAccess } from '@/lib/console/authorization'
import {
  getOrgMembers,
  createOrgInvitation,
  getUserByEmail,
  getOrgMemberRole,
} from '@/lib/console/storage'

/**
 * GET /api/orgs/[orgId]/members
 * List organization members
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgAccess(orgId)
    const members = await getOrgMembers(orgId)
    return NextResponse.json({ members })
  } catch (error) {
    return apiCatchError(error, 'Failed to list members')
  }
}

/**
 * POST /api/orgs/[orgId]/members
 * Invite a member by email (creates an invitation, role is always 'admin')
 * Body: { email: string }
 */
export async function POST(
  request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    const auth = await requireOrgAccess(orgId)

    const body = await request.json()
    const { email, role } = body as { email?: string; role?: string }

    if (!email || !email.trim()) {
      return apiError('Email is required', 400)
    }

    const memberRole: 'admin' | 'member' = role === 'admin' ? 'admin' : 'member'
    const emailLower = email.toLowerCase().trim()

    const existingUser = await getUserByEmail(emailLower)
    if (existingUser) {
      const existingRole = await getOrgMemberRole(orgId, existingUser.userId)
      if (existingRole) {
        return apiError('User is already a member of this organization', 409)
      }
    }

    const invitation = await createOrgInvitation(orgId, emailLower, auth.userId, memberRole)
    return NextResponse.json({ invitation }, { status: 201 })
  } catch (error) {
    return apiCatchError(error, 'Failed to invite member')
  }
}
