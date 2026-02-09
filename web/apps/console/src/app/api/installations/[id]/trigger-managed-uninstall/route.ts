import { NextResponse } from 'next/server'
import { requireOwnerPermission } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation } from '@/lib/console/storage'
import { triggerOCIInstallerJob } from '@/lib/console/aca-jobs'

const STALE_TIMEOUT_MS = 30 * 60 * 1000 // 30 minutes

export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireOwnerPermission(id)

    const installation = await getInstallationById(id)
    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }

    if (installation.cloudProvider !== 'oci') {
      return NextResponse.json({ error: 'Managed uninstall is only available for Kloudlite Cloud installations' }, { status: 400 })
    }

    // If a job is already running and not stale, return existing execution
    if (
      (installation.acaJobStatus === 'running' || installation.acaJobStatus === 'pending') &&
      installation.acaJobStartedAt &&
      Date.now() - new Date(installation.acaJobStartedAt).getTime() < STALE_TIMEOUT_MS
    ) {
      return NextResponse.json({
        success: true,
        executionName: installation.acaJobExecutionName,
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
      return NextResponse.json({ error: 'Kloudlite Cloud is not configured on this server' }, { status: 500 })
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
      acaJobExecutionName: result.executionName,
      acaJobStartedAt: new Date().toISOString(),
      acaJobCompletedAt: undefined,
      acaJobError: undefined,
    })

    return NextResponse.json({ success: true, executionName: result.executionName })
  } catch (error) {
    console.error('Error triggering managed uninstall:', error)
    const message = error instanceof Error ? error.message : 'Failed to trigger uninstall'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
