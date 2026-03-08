export const runtime = 'nodejs'

import { NextResponse } from 'next/server'
import { apiCatchError } from '@/lib/api-helpers'
import { requireOrgAccess } from '@/lib/console/authorization'
import { getOrgInvitations } from '@/lib/console/storage'

/**
 * GET /api/orgs/[orgId]/invitations
 * List pending invitations for an organization
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ orgId: string }> },
) {
  const { orgId } = await params

  try {
    await requireOrgAccess(orgId)
    const invitations = await getOrgInvitations(orgId)
    return NextResponse.json({ invitations })
  } catch (error) {
    return apiCatchError(error, 'Failed to list invitations')
  }
}
