import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationAccess } from '@/lib/console/authorization'
import {
  getInstallationById,
  updateInstallation,
  deleteInstallation,
  deleteDnsConfigurations,
  deleteDomainReservation,
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
    if (!installation.deployJobExecutionName) {
      // If there's no ACA execution AND no progress data, return 404
      if (!installation.deployJobStatus && !installation.deployJobOperation) {
        return apiError('No job execution found', 404)
      }

      // Auto-delete after successful uninstall (BYOC path)
      if (installation.deployJobOperation === 'uninstall' && installation.deployJobStatus === 'succeeded') {
        try {
          console.log(`[job-status] Auto-deleting installation ${id} after successful uninstall (BYOC)`)
        const dnsRecordIds = await deleteDnsConfigurations(id)
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
      status: installation.deployJobStatus || 'unknown',
      startedAt: installation.deployJobStartedAt,
      completedAt: installation.deployJobCompletedAt,
      error: installation.deployJobError,
      operation: installation.deployJobOperation,
      currentStep: installation.deployJobCurrentStep,
      totalSteps: installation.deployJobTotalSteps,
      stepDescription: installation.deployJobStepDescription,
      })
    }

    const result = await getJobExecutionStatus(installation.deployJobExecutionName)

    // Update installation if status changed
    if (result.status !== installation.deployJobStatus) {
      const updates: Record<string, string | undefined> = {
        deployJobStatus: result.status,
      }
      if (result.completedAt) {
        updates.deployJobCompletedAt = result.completedAt
      }
      if (result.error) {
        updates.deployJobError = result.error
      }
      await updateInstallation(id, updates)
    }

    // Auto-delete after successful uninstall
    if (installation.deployJobOperation === 'uninstall' && result.status === 'succeeded') {
      try {
        console.log(`[job-status] Auto-deleting installation ${id} after successful uninstall`)
        const dnsRecordIds = await deleteDnsConfigurations(id)
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
      executionName: installation.deployJobExecutionName,
      startedAt: result.startedAt || installation.deployJobStartedAt,
      completedAt: result.completedAt || installation.deployJobCompletedAt,
      error: result.error || installation.deployJobError,
      operation: updatedInstallation?.deployJobOperation || installation.deployJobOperation,
      currentStep: updatedInstallation?.deployJobCurrentStep ?? installation.deployJobCurrentStep,
      totalSteps: updatedInstallation?.deployJobTotalSteps ?? installation.deployJobTotalSteps,
      stepDescription: updatedInstallation?.deployJobStepDescription || installation.deployJobStepDescription,
    })
  } catch (error) {
    console.error('Error getting job status:', error)
    return apiCatchError(error, 'Failed to get job status')
  }
}
