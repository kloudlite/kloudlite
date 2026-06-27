import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationOwner } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation } from '@/lib/console/storage'
import { readFileSync } from 'fs'

export const runtime = 'nodejs'

function getServiceAccountToken(): string {
  return readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf8')
}

const STALE_TIMEOUT_MS = 30 * 60 * 1000

export async function POST(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireInstallationOwner(id)

    const installation = await getInstallationById(id)
    if (!installation) return apiError('Installation not found', 404)

    if (installation.cloudProvider !== 'oci') {
      return apiError('Managed uninstall is only available for Kloudlite Cloud installations', 400)
    }

    if (
      (installation.deployJobStatus === 'running' || installation.deployJobStatus === 'pending') &&
      installation.deployJobStartedAt &&
      Date.now() - new Date(installation.deployJobStartedAt).getTime() < STALE_TIMEOUT_MS
    ) {
      return NextResponse.json({ success: true, executionName: installation.deployJobExecutionName, message: 'Job already running' })
    }

    const ociTenancy = process.env.KLOUDLITE_OCI_TENANCY
    const ociUser = process.env.KLOUDLITE_OCI_USER
    const ociRegion = process.env.KLOUDLITE_OCI_REGION || 'ap-mumbai-1'
    const ociFingerprint = process.env.KLOUDLITE_OCI_FINGERPRINT
    const ociPrivateKey = process.env.KLOUDLITE_OCI_PRIVATE_KEY

    if (!ociTenancy || !ociUser || !ociFingerprint || !ociPrivateKey) {
      return apiError('Kloudlite Cloud is not configured on this server', 500)
    }

    const jobName = `oci-uninstall-${id.substring(0, 8)}`

    const k8sResponse = await fetch(
      `https://kubernetes.default.svc/apis/batch/v1/namespaces/production/jobs`,
      {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getServiceAccountToken()}`,
        },
        body: JSON.stringify({
          apiVersion: 'batch/v1',
          kind: 'Job',
          metadata: { name: jobName },
          spec: {
            ttlSecondsAfterFinished: 604800,
            template: {
              spec: {
                imagePullSecrets: [{ name: 'ghcr-pull-secret' }],
                containers: [{
                  name: 'oci-installer',
                  image: 'ghcr.io/kloudlite/kloudlite/oci-installer:20260626-fdf5f7d2',
                  env: [
                    { name: 'OPERATION', value: 'uninstall' },
                    { name: 'INSTALLATION_KEY', value: installation.installationKey },
                    { name: 'CONSOLE_BASE_URL', value: 'https://console.kloudlite.io' },
                    { name: 'OCI_CLI_TENANCY', value: ociTenancy },
                    { name: 'OCI_CLI_USER', value: ociUser },
                    { name: 'OCI_CLI_REGION', value: ociRegion },
                    { name: 'OCI_CLI_FINGERPRINT', value: ociFingerprint },
                    { name: 'OCI_CLI_KEY_CONTENT', valueFrom: { secretKeyRef: { name: 'console-secrets', key: 'KLOUDLITE_OCI_PRIVATE_KEY' } } },
                    { name: 'SKIP_LB', value: 'true' },
                    { name: 'ENABLE_DELETION_PROTECTION', value: 'true' },
                  ],
                  resources: { requests: { cpu: '1', memory: '2Gi' }, limits: { cpu: '2', memory: '4Gi' } },
                }],
                restartPolicy: 'Never',
              },
            },
            backoffLimit: 0,
          },
        }),
      },
    )

    if (!k8sResponse.ok) {
      const errText = await k8sResponse.text()
      throw new Error(`K8s API error: ${errText}`)
    }

    await updateInstallation(id, {
      deployJobExecutionName: jobName,
      deployJobStatus: 'running',
      deployJobStartedAt: new Date().toISOString(),
      deployJobError: undefined,
      deployJobOperation: 'uninstall',
      deployJobCurrentStep: 0,
      deployJobTotalSteps: 4,
      deployJobStepDescription: 'Starting uninstallation...',
    })

    return NextResponse.json({ success: true, executionName: jobName })
  } catch (error) {
    return apiCatchError(error, 'Failed to trigger uninstall')
  }
}
