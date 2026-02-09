/**
 * Azure Container Apps Jobs helper
 *
 * Manages OCI installer job executions on Azure Container Apps.
 * Uses DefaultAzureCredential for auth (managed identity in production,
 * service principal env vars in development).
 */

import { ContainerAppsAPIClient } from '@azure/arm-appcontainers'
import { DefaultAzureCredential } from '@azure/identity'

const SUBSCRIPTION_ID = process.env.AZURE_SUBSCRIPTION_ID || ''
const RESOURCE_GROUP = process.env.AZURE_RESOURCE_GROUP || 'rg-kloudlite'
const JOB_NAME = process.env.OCI_INSTALLER_JOB_NAME || 'job-oci-installer'
const OCI_INSTALLER_IMAGE = process.env.OCI_INSTALLER_IMAGE || 'ghcr.io/kloudlite/kloudlite/oci-installer:sha-d5c9653'

function getClient(): ContainerAppsAPIClient {
  const credential = new DefaultAzureCredential()
  return new ContainerAppsAPIClient(credential, SUBSCRIPTION_ID)
}

export type JobStatus = 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown'

export interface TriggerJobParams {
  operation: 'install' | 'uninstall'
  installationKey: string
  consoleBaseURL?: string
  ociTenancy: string
  ociUser: string
  ociRegion: string
  ociCompartment?: string
  ociFingerprint: string
  ociPrivateKey: string
  skipLB?: boolean
  enableDeletionProtection?: boolean
}

export interface JobExecutionResult {
  executionName: string
  status: JobStatus
}

export interface JobStatusResult {
  status: JobStatus
  startedAt?: string
  completedAt?: string
  error?: string
}

/**
 * Trigger an OCI installer job execution with per-execution env var overrides.
 *
 * Uses beginStart (non-waiting) to return immediately with the execution name.
 * The caller should poll getJobExecutionStatus for progress.
 */
export async function triggerOCIInstallerJob(
  params: TriggerJobParams,
): Promise<JobExecutionResult> {
  const client = getClient()

  const envVars = [
    { name: 'OPERATION', value: params.operation },
    { name: 'INSTALLATION_KEY', value: params.installationKey },
    { name: 'CONSOLE_BASE_URL', value: params.consoleBaseURL || 'https://console.kloudlite.io' },
    { name: 'OCI_CLI_TENANCY', value: params.ociTenancy },
    { name: 'OCI_CLI_USER', value: params.ociUser },
    { name: 'OCI_CLI_REGION', value: params.ociRegion },
    { name: 'OCI_CLI_COMPARTMENT', value: params.ociCompartment || '' },
    { name: 'OCI_CLI_FINGERPRINT', value: params.ociFingerprint },
    { name: 'OCI_CLI_KEY_CONTENT', value: params.ociPrivateKey.replace(/\\n/g, '\n') },
    { name: 'SKIP_LB', value: String(params.skipLB ?? false) },
    { name: 'ENABLE_DELETION_PROTECTION', value: String(params.enableDeletionProtection ?? true) },
  ]

  // Use beginStart (non-blocking) so we return the execution name immediately
  const poller = await client.jobs.beginStart(RESOURCE_GROUP, JOB_NAME, {
    template: {
      containers: [
        {
          name: 'oci-installer',
          image: OCI_INSTALLER_IMAGE,
          env: envVars,
        },
      ],
    },
  })

  // Get the initial result which contains the execution name
  const initialResult = poller.getResult()
  const executionName = initialResult?.name || ''

  return {
    executionName,
    status: 'running',
  }
}

/**
 * Get the status of a job execution.
 * Uses the client.jobExecution() method to get full execution details including status.
 */
export async function getJobExecutionStatus(executionName: string): Promise<JobStatusResult> {
  const client = getClient()

  const execution = await client.jobExecution(RESOURCE_GROUP, JOB_NAME, executionName)

  console.log(`[aca-jobs] execution ${executionName} raw status: ${execution.status}`)
  const status = mapRunningState(execution.status)

  return {
    status,
    startedAt: execution.startTime?.toISOString(),
    completedAt: execution.endTime?.toISOString(),
  }
}

function mapRunningState(state?: string): JobStatus {
  if (!state) return 'unknown'
  const s = state.toLowerCase()
  if (s === 'succeeded') return 'succeeded'
  if (s === 'failed') return 'failed'
  if (s === 'running' || s === 'processing') return 'running'
  if (s === 'stopped' || s === 'degraded') return 'failed'
  return 'unknown'
}
