import { NextResponse } from 'next/server'
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
    return NextResponse.json({ error: 'Not authenticated' }, { status: 401 })
  }

  const { id } = await params

  // Verify user has access to this installation
  const role = await getMemberRole(id, session.user.id)
  const installation = await getInstallationById(id)
  if (!role && installation?.userId !== session.user.id) {
    return NextResponse.json({ error: 'Forbidden' }, { status: 403 })
  }

  const [subscriptions, plans] = await Promise.all([
    getSubscriptionsByInstallation(id),
    getPlans(),
  ])

  return NextResponse.json({ subscriptions, plans })
}
