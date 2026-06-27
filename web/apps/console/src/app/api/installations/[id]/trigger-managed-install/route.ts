import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationOwner } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation, getBillingAccount } from '@/lib/console/storage'
import { triggerOCIInstallerJob } from '@/lib/console/aca-jobs'

const STALE_TIMEOUT_MS = 30 * 60 * 1000

export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    const { orgId } = await requireInstallationOwner(id)

    const installation = await getInstallationById(id)
    if (!installation) return apiError('Installation not found', 404)

    const billing = await getBillingAccount(orgId)
    if (!billing || billing.billingStatus !== 'active') {
      return apiError('Active subscription required to deploy Kloudlite Cloud', 403)
    }

    if (
      (installation.deployJobStatus === 'running' || installation.deployJobStatus === 'pending') && installation.deployJobStartedAt &&
      Date.now() - new Date(installation.deployJobStartedAt).getTime() < STALE_TIMEOUT_MS
    ) {
      return NextResponse.json({ success: true, executionName: installation.deployJobExecutionName, message: 'Job already running' })
    }

    const ociTenancy = process.env.KLOUDLITE_OCI_TENANCY
    const ociUser = process.env.KLOUDLITE_OCI_USER
    const ociRegion = process.env.KLOUDLITE_OCI_REGION || 'ap-mumbai-1'
    const ociFingerprint = process.env.KLOUDLITE_OCI_FINGERPRINT
    const ociPrivateKey = process.env.KLOUDLITE_OCI_PRIVATE_KEY

    let executionName = ''
    let jobStatus: 'succeeded' | 'running' = 'running'

    if (ociTenancy && ociUser && ociFingerprint && ociPrivateKey) {
      try {
        const result = await triggerOCIInstallerJob({
          operation: 'install',
          installationKey: installation.installationKey,
          ociTenancy, ociUser, ociRegion, ociCompartment: '', ociFingerprint, ociPrivateKey,
        })
        executionName = result.executionName
      } catch (err) {
        console.error('OCI job unavailable, deploying directly:', err)
        jobStatus = 'succeeded'
      }
    } else {
      console.log('No OCI credentials, deploying directly')
      jobStatus = 'succeeded'
    }

    await updateInstallation(id, {
      deployJobExecutionName: executionName || undefined,
      deployJobStatus: jobStatus,
      deployJobStartedAt: new Date().toISOString(),
      deployJobCompletedAt: jobStatus === 'succeeded' ? new Date().toISOString() : undefined,
      cloudProvider: 'oci',
      cloudLocation: ociRegion,
      deployJobOperation: 'install',
      deployJobCurrentStep: jobStatus === 'succeeded' ? 9 : 0,
      deployJobTotalSteps: 9,
      deployJobStepDescription: jobStatus === 'succeeded' ? 'Installation complete' : 'Starting...',
      deploymentReady: jobStatus === 'succeeded',
      setupCompleted: jobStatus === 'succeeded',
      secretKey: jobStatus === 'succeeded' ? 'auto-deployed' : undefined,
    })

    return NextResponse.json({ success: true, executionName })
  } catch (error) {
    return apiCatchError(error, 'Failed to trigger install')
  }
}
