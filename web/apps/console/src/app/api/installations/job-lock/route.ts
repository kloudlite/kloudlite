import { NextResponse } from 'next/server'
import { getInstallationByKey, updateInstallation } from '@/lib/console/storage'

export const runtime = 'nodejs'

/**
 * POST /api/installations/job-lock
 * Body: { installationKey: string, action: "lock" | "unlock" }
 *
 * Called by the OCI installer job to acquire/release a lock.
 * Lock is per installation key — only one job can run at a time.
 */
export async function POST(request: Request) {
  try {
    const { installationKey, action } = await request.json()

    if (!installationKey || !action) {
      return NextResponse.json({ error: 'installationKey and action are required' }, { status: 400 })
    }

    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }

    if (action === 'lock') {
      // Only reject if another job is actively running (not just pending/triggered)
      if (
        installation.acaJobStatus === 'running' &&
        installation.acaJobStartedAt &&
        Date.now() - new Date(installation.acaJobStartedAt).getTime() < 30 * 60 * 1000
      ) {
        return NextResponse.json({ acquired: false, message: 'Job already running' })
      }

      await updateInstallation(installation.id, {
        acaJobStatus: 'running',
        acaJobStartedAt: new Date().toISOString(),
        acaJobError: undefined,
      })

      return NextResponse.json({ acquired: true })
    }

    if (action === 'unlock') {
      await updateInstallation(installation.id, {
        acaJobStatus: 'succeeded',
        acaJobCompletedAt: new Date().toISOString(),
      })

      return NextResponse.json({ released: true })
    }

    return NextResponse.json({ error: 'action must be "lock" or "unlock"' }, { status: 400 })
  } catch (error) {
    console.error('Job lock error:', error)
    return NextResponse.json({ error: 'Internal server error' }, { status: 500 })
  }
}
