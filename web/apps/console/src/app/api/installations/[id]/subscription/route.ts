import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getSubscriptionsByInstallation,
  getPlans,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'

export const runtime = 'nodejs'

export async function GET(
  _request: Request,
  { params }: { params: Promise<{ id: string }> },
) {
  const session = await getRegistrationSession()
  if (!session?.user) {
    return apiError('Not authenticated', 401)
  }

  const { id } = await params

  // Verify user has access to this installation
  const role = await getMemberRole(id, session.user.id)
  const installation = await getInstallationById(id)
  if (!role && installation?.userId !== session.user.id) {
    return apiError('Forbidden', 403)
  }

  const [subscriptions, plans] = await Promise.all([
    getSubscriptionsByInstallation(id),
    getPlans(),
  ])

  return NextResponse.json({ subscriptions, plans })
}
