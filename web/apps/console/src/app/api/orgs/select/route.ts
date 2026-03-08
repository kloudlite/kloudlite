import { NextRequest, NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import { isOrgMember } from '@/lib/console/storage'
import { cookies } from 'next/headers'

/**
 * POST /api/orgs/select
 * Set the selected organization cookie
 */
export async function POST(request: NextRequest) {
  const session = await getRegistrationSession()
  if (!session?.user) {
    return apiError('Unauthorized', 401)
  }

  const { orgId } = await request.json()
  if (!orgId) {
    return apiError('orgId is required', 400)
  }

  // Verify user is a member of the org
  const isMember = await isOrgMember(orgId, session.user.id)
  if (!isMember) {
    return apiError('Not a member of this organization', 403)
  }

  const cookieStore = await cookies()
  cookieStore.set('selected_org_id', orgId, {
    httpOnly: true,
    secure: process.env.NODE_ENV !== 'development',
    sameSite: 'lax',
    maxAge: 60 * 60 * 24 * 365, // 1 year
    path: '/',
  })

  return NextResponse.json({ success: true })
}
