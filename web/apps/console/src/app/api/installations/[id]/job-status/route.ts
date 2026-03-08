import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationAccess } from '@/lib/console/authorization'
import {
  getInstallationById,
  updateInstallation,
  deleteInstallation,
  deleteIpRecords,
  deleteDomainReservation,
  cancelStripeSubscriptionForInstallation,
} from '@/lib/console/storage'
import { deleteDnsRecords } from '@/lib/console/cloudflare-dns'
import { getJobExecutionStatus } from '@/lib/console/aca-jobs'

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireInstallationAccess(id)

    const installation = await getInstallationById(id)
    if (!installation) {
      return apiError('Installation not found', 404)
    }

    // For BYOC installations (AWS/GCP/Azure), there's no ACA job execution.
    // Progress is reported via the job-progress endpoint and stored directly in the DB.
    if (!installation.acaJobExecutionName) {
      // If there's no ACA execution AND no progress data, return 404
      if (!installation.acaJobStatus && !installation.acaJobOperation) {
        return apiError('No job execution found', 404)
      }

      // Auto-delete after successful uninstall (BYOC path)
      if (installation.acaJobOperation === 'uninstall' && installation.acaJobStatus === 'succeeded') {
        try {
          console.log(`[job-status] Auto-deleting installation ${id} after successful uninstall (BYOC)`)
          await cancelStripeSubscriptionForInstallation(id)
          const dnsRecordIds = await deleteIpRecords(id)
          if (dnsRecordIds.length > 0) {
            await deleteDnsRecords(dnsRecordIds)
          }
          await deleteDomainReservation(id)
          await deleteInstallation(id)
          console.log(`[job-status] Installation ${id} auto-deleted (BYOC)`)
        } catch (deleteErr) {
          console.error(`[job-status] Failed to auto-delete installation ${id}:`, deleteErr)
        }
        return NextResponse.json({ status: 'succeeded', operation: 'uninstall', deleted: true })
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

    // Auto-delete after successful uninstall
    if (installation.acaJobOperation === 'uninstall' && result.status === 'succeeded') {
      try {
        console.log(`[job-status] Auto-deleting installation ${id} after successful uninstall`)
        await cancelStripeSubscriptionForInstallation(id)
        const dnsRecordIds = await deleteIpRecords(id)
        if (dnsRecordIds.length > 0) {
          await deleteDnsRecords(dnsRecordIds)
        }
        await deleteDomainReservation(id)
        await deleteInstallation(id)
        console.log(`[job-status] Installation ${id} auto-deleted`)
      } catch (deleteErr) {
        console.error(`[job-status] Failed to auto-delete installation ${id}:`, deleteErr)
      }
      return NextResponse.json({ status: 'succeeded', operation: 'uninstall', deleted: true })
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
    return apiCatchError(error, 'Failed to get job status')
  }
}
