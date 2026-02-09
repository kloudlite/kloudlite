import { NextResponse } from 'next/server'
import { requireOwnerPermission } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation } from '@/lib/console/storage'
import { triggerOCIInstallerJob } from '@/lib/console/aca-jobs'

export async function POST(request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireOwnerPermission(id)

    const installation = await getInstallationById(id)
    if (!installation) {
      return NextResponse.json({ error: 'Installation not found' }, { status: 404 })
    }

    const body = await request.json()
    const { ociTenancy, ociUser, ociRegion, ociCompartment, ociFingerprint, ociPrivateKey } = body

    if (!ociTenancy || !ociUser || !ociRegion || !ociFingerprint || !ociPrivateKey) {
      return NextResponse.json(
        { error: 'Missing required OCI credentials' },
        { status: 400 },
      )
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
      acaJobStatus: result.status,
      acaJobStartedAt: new Date().toISOString(),
      acaJobCompletedAt: undefined,
      acaJobError: undefined,
    })

    return NextResponse.json({
      success: true,
      executionName: result.executionName,
    })
  } catch (error) {
    console.error('Error triggering OCI uninstall:', error)
    const message = error instanceof Error ? error.message : 'Failed to trigger uninstall'
    const status = message.includes('Unauthorized') ? 401 : message.includes('Forbidden') ? 403 : 500
    return NextResponse.json({ error: message }, { status })
  }
}
