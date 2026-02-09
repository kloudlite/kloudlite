import { NextResponse } from 'next/server'
import { requireOwnerPermission } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation } from '@/lib/console/storage'
import { triggerOCIInstallerJob } from '@/lib/console/aca-jobs'

export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireOwnerPermission(id)

    const installation = await getInstallationById(id)
    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }

    if (!installation.subdomain) {
      return NextResponse.json(
        { error: 'No subdomain configured for this installation' },
        { status: 400 },
      )
    }

    // Read OCI credentials from server-side env vars
    const ociTenancy = process.env.KLOUDLITE_OCI_TENANCY
    const ociUser = process.env.KLOUDLITE_OCI_USER
    const ociRegion = process.env.KLOUDLITE_OCI_REGION
    const ociCompartment = process.env.KLOUDLITE_OCI_COMPARTMENT || ''
    const ociFingerprint = process.env.KLOUDLITE_OCI_FINGERPRINT
    const ociPrivateKey = process.env.KLOUDLITE_OCI_PRIVATE_KEY

    if (!ociTenancy || !ociUser || !ociRegion || !ociFingerprint || !ociPrivateKey) {
      console.error('Missing KLOUDLITE_OCI_* env vars for managed install')
      return NextResponse.json(
        { error: 'Kloudlite Cloud is not configured on this server' },
        { status: 500 },
      )
    }

    const result = await triggerOCIInstallerJob({
      operation: 'install',
      installationKey: installation.installationKey,
      ociTenancy,
      ociUser,
      ociRegion,
      ociCompartment,
      ociFingerprint,
      ociPrivateKey,
    })

    // Store execution info on the installation
    await updateInstallation(id, {
      acaJobExecutionName: result.executionName,
      acaJobStatus: result.status,
      acaJobStartedAt: new Date().toISOString(),
      acaJobCompletedAt: undefined,
      acaJobError: undefined,
      cloudProvider: 'oci',
      cloudLocation: ociRegion,
    })

    return NextResponse.json({
      success: true,
      executionName: result.executionName,
    })
  } catch (error) {
    console.error('Error triggering managed install:', error)
    const message = error instanceof Error ? error.message : 'Failed to trigger install'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
