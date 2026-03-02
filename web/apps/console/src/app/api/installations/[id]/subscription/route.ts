import { NextResponse } from 'next/server'
import { getRegistrationSession } from '@/lib/console-auth'
import { getSubscriptionsByInstallation, getPlans } from '@/lib/console/storage'

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

  const [subscriptions, plans] = await Promise.all([
    getSubscriptionsByInstallation(id),
    getPlans(),
  ])

  return NextResponse.json({ subscriptions, plans })
}
