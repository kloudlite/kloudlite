import { NextResponse } from 'next/server'
import { apiCatchError } from '@/lib/api-helpers'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationMembers } from '@/lib/console/storage'

/**
 * GET /api/installations/[id]/team/members
 * List all team members
 */
export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params

  try {
    await requireInstallationAccess(id)
    const members = await getInstallationMembers(id)

    return NextResponse.json({ members })
  } catch (error) {
    return apiCatchError(error, 'Failed to get members')
  }
}
