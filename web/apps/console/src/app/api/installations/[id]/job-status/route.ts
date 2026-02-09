import { NextResponse } from 'next/server'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation } from '@/lib/console/storage'
import { getJobExecutionStatus } from '@/lib/console/aca-jobs'

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireInstallationAccess(id)

    const installation = await getInstallationById(id)
    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }

    if (!installation.acaJobExecutionName) {
      return NextResponse.json({ error: 'No job execution found' }, { status: 404 })
    }

    const result = await getJobExecutionStatus(installation.acaJobExecutionName)

    // Update installation if status changed
    if (result.status !== installation.acaJobStatus) {
      const updates: Record<string, string | undefined> = {
        acaJobStatus: result.status,
      }
      if (result.completedAt) {
        updates.acaJobCompletedAt = result.completedAt
      }
      if (result.error) {
        updates.acaJobError = result.error
      }
      await updateInstallation(id, updates)
    }

    return NextResponse.json({
      status: result.status,
      executionName: installation.acaJobExecutionName,
      startedAt: result.startedAt || installation.acaJobStartedAt,
      completedAt: result.completedAt || installation.acaJobCompletedAt,
      error: result.error || installation.acaJobError,
    })
  } catch (error) {
    console.error('Error getting job status:', error)
    const message = error instanceof Error ? error.message : 'Failed to get job status'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
