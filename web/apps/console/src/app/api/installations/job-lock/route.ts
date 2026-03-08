import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import {
  getInstallationByKey,
  updateInstallation,
  deleteInstallation,
  deleteDnsConfigurations,
  deleteDomainReservation,
} from '@/lib/console/storage'
import { deleteDnsRecords } from '@/lib/console/cloudflare-dns'

export const runtime = 'nodejs'

/**
 * POST /api/installations/job-lock
 * Body: { installationKey: string, action: "lock" | "unlock", status?: "failed" }
 *
 * Called by the OCI installer job to acquire/release a lock.
 * Lock is per installation key — only one job can run at a time.
 */
export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { installationKey, action, status: jobStatus } = body

    if (!installationKey || !action) {
      return apiError('installationKey and action are required', 400)
    }

    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return apiError('Installation not found', 404)
    }

    if (action === 'lock') {
      // Only reject if another job is actively running (not just pending/triggered)
      if (
        installation.deployJobStatus === 'running' &&
        installation.deployJobStartedAt &&
        Date.now() - new Date(installation.deployJobStartedAt).getTime() < 30 * 60 * 1000
      ) {
        return NextResponse.json({ acquired: false, message: 'Job already running' })
      }

      await updateInstallation(installation.id, {
        deployJobStatus: 'running',
        deployJobStartedAt: new Date().toISOString(),
        deployJobError: undefined,
      })

      return NextResponse.json({ acquired: true })
    }

    if (action === 'unlock') {
      const finalStatus = jobStatus === 'failed' ? 'failed' : 'succeeded'
      const updates: Record<string, unknown> = {
        deployJobStatus: finalStatus,
        deployJobCompletedAt: new Date().toISOString(),
      }

      // Clear job fields after successful install (no longer needed)
      if (installation.deployJobOperation === 'install' && finalStatus === 'succeeded') {
        updates.deployJobOperation = null
        updates.deployJobCurrentStep = null
        updates.deployJobTotalSteps = null
        updates.deployJobStepDescription = null
      }

      await updateInstallation(installation.id, updates)

      // Auto-delete installation after successful uninstall
      if (installation.deployJobOperation === 'uninstall' && finalStatus === 'succeeded') {
        try {
          console.log(`Auto-deleting installation ${installation.id} after successful uninstall`)
          const dnsRecordIds = await deleteDnsConfigurations(installation.id)
          if (dnsRecordIds.length > 0) {
            await deleteDnsRecords(dnsRecordIds)
          }
          await deleteDomainReservation(installation.id)
          await deleteInstallation(installation.id)
          console.log(`Installation ${installation.id} auto-deleted successfully`)
        } catch (deleteErr) {
          console.error(`Failed to auto-delete installation ${installation.id}:`, deleteErr)
        }
      }

      return NextResponse.json({ released: true })
    }

    return apiError('action must be "lock" or "unlock"', 400)
  } catch (error) {
    console.error('Job lock error:', error)
    return apiError('Internal server error', 500)
  }
}
