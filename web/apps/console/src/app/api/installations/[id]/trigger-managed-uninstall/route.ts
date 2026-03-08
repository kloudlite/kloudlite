import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationOwner } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation } from '@/lib/console/storage'
import { triggerOCIInstallerJob } from '@/lib/console/aca-jobs'

const STALE_TIMEOUT_MS = 30 * 60 * 1000 // 30 minutes

export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireInstallationOwner(id)

    const installation = await getInstallationById(id)
    if (!installation) {
      return apiError('Installation not found', 404)
    }

    if (installation.cloudProvider !== 'oci') {
      return apiError('Managed uninstall is only available for Kloudlite Cloud installations', 400)
    }

    // If a job is already running and not stale, return existing execution
    if (
      (installation.deployJobStatus === 'running' || installation.deployJobStatus === 'pending') &&
      installation.deployJobStartedAt &&
      Date.now() - new Date(installation.deployJobStartedAt).getTime() < STALE_TIMEOUT_MS
    ) {
      return NextResponse.json({
        success: true,
        executionName: installation.deployJobExecutionName,
        message: 'Job already running',
      })
    }

    const ociTenancy = process.env.KLOUDLITE_OCI_TENANCY
    const ociUser = process.env.KLOUDLITE_OCI_USER
    const ociRegion = process.env.KLOUDLITE_OCI_REGION
    const ociCompartment = process.env.KLOUDLITE_OCI_COMPARTMENT || ''
    const ociFingerprint = process.env.KLOUDLITE_OCI_FINGERPRINT
    const ociPrivateKey = process.env.KLOUDLITE_OCI_PRIVATE_KEY

    if (!ociTenancy || !ociUser || !ociRegion || !ociFingerprint || !ociPrivateKey) {
      console.error('Missing KLOUDLITE_OCI_* env vars for managed uninstall')
      return apiError('Kloudlite Cloud is not configured on this server', 500)
    }

    const result = await triggerOCIInstallerJob({
      operation: 'uninstall',
      installationKey: installation.installationKey,
      ociTenancy,
      ociUser,
      ociRegion,
      ociCompartment,
      ociFingerprint,
      ociPrivateKey,
    })

    await updateInstallation(id, {
      deployJobExecutionName: result.executionName,
      deployJobStatus: 'running',
      deployJobStartedAt: new Date().toISOString(),
      deployJobCompletedAt: undefined,
      deployJobError: undefined,
      deployJobOperation: 'uninstall',
      deployJobCurrentStep: 0,
      deployJobTotalSteps: 4,
      deployJobStepDescription: 'Starting uninstallation...',
    })

    return NextResponse.json({ success: true, executionName: result.executionName })
  } catch (error) {
    console.error('Error triggering managed uninstall:', error)
    return apiCatchError(error, 'Failed to trigger uninstall')
  }
}
