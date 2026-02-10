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

    // For BYOC installations (AWS/GCP/Azure), there's no ACA job execution.
    // Progress is reported via the job-progress endpoint and stored directly in the DB.
    if (!installation.acaJobExecutionName) {
      // If there's no ACA execution AND no progress data, return 404
      if (!installation.acaJobStatus && !installation.acaJobOperation) {
        return NextResponse.json({ error: 'No job execution found' }, { status: 404 })
      }

      // Return DB-stored progress for BYOC installations
      return NextResponse.json({
        status: installation.acaJobStatus || 'unknown',
        startedAt: installation.acaJobStartedAt,
        completedAt: installation.acaJobCompletedAt,
        error: installation.acaJobError,
        operation: installation.acaJobOperation,
        currentStep: installation.acaJobCurrentStep,
        totalSteps: installation.acaJobTotalSteps,
        stepDescription: installation.acaJobStepDescription,
      })
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

    // Re-fetch installation to get latest progress fields (may have been updated by job-progress callback)
    const updatedInstallation = await getInstallationById(id)

    return NextResponse.json({
      status: result.status,
      executionName: installation.acaJobExecutionName,
      startedAt: result.startedAt || installation.acaJobStartedAt,
      completedAt: result.completedAt || installation.acaJobCompletedAt,
      error: result.error || installation.acaJobError,
      operation: updatedInstallation?.acaJobOperation || installation.acaJobOperation,
      currentStep: updatedInstallation?.acaJobCurrentStep ?? installation.acaJobCurrentStep,
      totalSteps: updatedInstallation?.acaJobTotalSteps ?? installation.acaJobTotalSteps,
      stepDescription: updatedInstallation?.acaJobStepDescription || installation.acaJobStepDescription,
    })
  } catch (error) {
    console.error('Error getting job status:', error)
    const message = error instanceof Error ? error.message : 'Failed to get job status'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
