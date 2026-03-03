import { NextResponse } from 'next/server'
import { apiError } from '@/lib/api-helpers'
import { getInstallationByKey, updateInstallation } from '@/lib/console/storage'

export const runtime = 'nodejs'

/**
 * POST /api/installations/job-progress
 * Body: { installationKey, operation, currentStep, totalSteps, stepDescription }
 *
 * Called by the installer (OCI job or BYOC kli CLI) at each step to report progress.
 * Fire-and-forget from the Go side — errors are non-fatal.
 *
 * Also sets acaJobStatus to 'running' so the frontend detects active jobs.
 * For BYOC installations (AWS/GCP/Azure), this is the only way the job status gets set
 * since they don't use the trigger-managed-install/uninstall endpoints.
 */
export async function POST(request: Request) {
  try {
    const body = await request.json()
    const { installationKey, operation, currentStep, totalSteps, stepDescription, completed } = body

    if (!installationKey || !operation || currentStep == null || !totalSteps) {
      return apiError('installationKey, operation, currentStep, and totalSteps are required', 400)
    }

    const installation = await getInstallationByKey(installationKey)
    if (!installation) {
      return apiError('Installation not found', 404)
    }

    // Set acaJobStatus to 'running' so the frontend shows INSTALLING/UNINSTALLING badges.
    // For BYOC (AWS/GCP/Azure), this is the only signal that a job is active.
    // When completed=true, mark job as succeeded.
    const updates: Record<string, unknown> = {
      acaJobOperation: operation,
      acaJobCurrentStep: currentStep,
      acaJobTotalSteps: totalSteps,
      acaJobStepDescription: stepDescription || '',
      acaJobStatus: completed ? 'succeeded' : 'running',
    }

    // If this is the first progress report, record the start time and reset deploymentReady
    if (!installation.acaJobStartedAt || installation.acaJobStatus !== 'running') {
      updates.acaJobStartedAt = new Date().toISOString()
      if (operation === 'install') {
        updates.deploymentReady = false
      }
    }

    // Record completion time and clear job fields after successful install
    if (completed) {
      updates.acaJobCompletedAt = new Date().toISOString()
      if (operation === 'install') {
        updates.acaJobOperation = null
        updates.acaJobCurrentStep = null
        updates.acaJobTotalSteps = null
        updates.acaJobStepDescription = null
      }
    }

    await updateInstallation(installation.id, updates)

    return NextResponse.json({ ok: true })
  } catch (error) {
    console.error('Job progress error:', error)
    return apiError('Internal server error', 500)
  }
}
