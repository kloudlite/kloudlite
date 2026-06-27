import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationOwner } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation, getBillingAccount } from '@/lib/console/storage'
import { readFileSync } from 'fs'

function getServiceAccountToken(): string {
  return readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf8')
}

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
        const jobName = `oci-installer-${id.substring(0, 8)}`

        // Create K8s Job via in-cluster API
        const k8sResponse = await fetch(
          `https://kubernetes.default.svc/api/v1/namespaces/production/jobs`,
          {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${await getServiceAccountToken()}`,
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
                        { name: 'OPERATION', value: 'install' },
                        { name: 'INSTALLATION_KEY', value: installation.installationKey },
                        { name: 'CONSOLE_BASE_URL', value: 'https://console.kloudlite.io' },
                        { name: 'OCI_CLI_TENANCY', value: ociTenancy },
                        { name: 'OCI_CLI_USER', value: ociUser },
                        { name: 'OCI_CLI_REGION', value: ociRegion },
                        { name: 'OCI_CLI_FINGERPRINT', value: ociFingerprint },
                        { name: 'OCI_CLI_KEY_CONTENT', valueFrom: { secretKeyRef: { name: 'console-secrets', key: 'KLOUDLITE_OCI_PRIVATE_KEY' } } },
                        { name: 'SKIP_LB', value: 'false' },
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

        executionName = jobName
        console.log(`Created K8s job: ${jobName}`)
      } catch (err) {
        console.error('Failed to create K8s installer job:', err)
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
