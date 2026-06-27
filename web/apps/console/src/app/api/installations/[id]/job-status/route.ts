import { NextResponse } from 'next/server'
import { apiError, apiCatchError } from '@/lib/api-helpers'
import { requireInstallationAccess } from '@/lib/console/authorization'
import { getInstallationById, updateInstallation } from '@/lib/console/storage'
import { readFileSync } from 'fs'

export const runtime = 'nodejs'

function getServiceAccountToken(): string {
  return readFileSync('/var/run/secrets/kubernetes.io/serviceaccount/token', 'utf8')
}

async function getK8sJobStatus(jobName: string): Promise<{ status: string; error?: string }> {
  try {
    const token = getServiceAccountToken()
    const res = await fetch(
      `https://kubernetes.default.svc/apis/batch/v1/namespaces/production/jobs/${jobName}`,
      { headers: { Authorization: `Bearer ${token}` } },
    )
    if (!res.ok) return { status: 'unknown' }

    const job = await res.json()
    const conditions = job?.status?.conditions || []

    if (conditions.some((c: any) => c.type === 'Complete' && c.status === 'True')) {
      return { status: 'succeeded' }
    }
    if (conditions.some((c: any) => c.type === 'Failed' && c.status === 'True')) {
      const reason = conditions.find((c: any) => c.type === 'Failed')?.reason || 'Job failed'
      return { status: 'failed', error: reason }
    }

    // Check pod status for detailed progress
    const podRes = await fetch(
      `https://kubernetes.default.svc/api/v1/namespaces/production/pods?labelSelector=job-name=${jobName}`,
      { headers: { Authorization: `Bearer ${token}` } },
    )
    if (podRes.ok) {
      const podList = await podRes.json()
      const pod = podList?.items?.[0]
      if (pod?.status?.containerStatuses?.[0]?.state?.waiting?.reason === 'ImagePullBackOff') {
        return { status: 'failed', error: 'Failed to pull OCI installer image' }
      }
      if (pod?.status?.phase === 'Running') {
        return { status: 'running' }
      }
      if (pod?.status?.phase === 'Pending') {
        return { status: 'pending' }
      }
      if (pod?.status?.phase === 'Succeeded') {
        return { status: 'succeeded' }
      }
      if (pod?.status?.phase === 'Failed') {
        return { status: 'failed', error: 'Job execution failed' }
      }
    }

    return { status: 'running' }
  } catch {
    return { status: 'unknown' }
  }
}

export async function GET(_request: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params

  try {
    await requireInstallationAccess(id)

    const installation = await getInstallationById(id)
    if (!installation) return apiError('Installation not found', 404)

    // For OCI managed installations, check the K8s Job status
    if (installation.cloudProvider === 'oci' && installation.deployJobExecutionName) {
      const jobStatus = await getK8sJobStatus(installation.deployJobExecutionName)

      if (jobStatus.status !== installation.deployJobStatus) {
        const updates: Record<string, any> = { deployJobStatus: jobStatus.status }
        if (jobStatus.status === 'succeeded') {
          updates.deploymentReady = true
          updates.setupCompleted = true
          updates.secretKey = 'auto-deployed'
          updates.deployJobCompletedAt = new Date().toISOString()
          updates.deployJobCurrentStep = 9
          updates.deployJobStepDescription = 'Installation complete'
        }
        if (jobStatus.status === 'failed') {
          updates.deployJobError = jobStatus.error
          updates.deployJobCompletedAt = new Date().toISOString()
        }
        await updateInstallation(id, updates)
      }

      return NextResponse.json({
        status: jobStatus.status,
        operation: installation.deployJobOperation,
        error: jobStatus.error || installation.deployJobError,
        executionName: installation.deployJobExecutionName,
      })
    }

    // For installations without a K8s job, return current DB status
    return NextResponse.json({
      status: installation.deployJobStatus || 'unknown',
      operation: installation.deployJobOperation,
      error: installation.deployJobError,
    })
  } catch (error) {
    return apiCatchError(error, 'Failed to get job status')
  }
}
