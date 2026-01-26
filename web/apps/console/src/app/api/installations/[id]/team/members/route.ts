import { NextResponse } from 'next/server'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationMembers } from '@/lib/console/supabase-storage-service'

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
    const message = error instanceof Error ? error.message : 'Failed to get members'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
