/**
 * Installation Management Functions
 * Installations belong to organizations (org_id), not users directly.
 */

import type { Database } from '../supabase-types'
import { supabase } from '../supabase'
import type { Installation, InstallationRow, DnsConfigurationRow } from './types'

function mapToInstallation(data: InstallationRow, dnsConfigurations: DnsConfigurationRow[] = []): Installation {
  return {
    id: data.id,
    orgId: data.org_id,
    name: data.name || undefined,
    description: data.description || undefined,
    installationKey: data.installation_key,
    secretKey: data.secret_key || undefined,
    setupCompleted: data.setup_completed,
    subdomain: data.subdomain || undefined,
    reservedAt: data.reserved_at || undefined,
    deploymentReady: data.deployment_ready || undefined,
    lastHealthCheck: data.last_health_check || undefined,
    cloudProvider: data.cloud_provider || undefined,
    cloudLocation: data.cloud_location || undefined,
    deployJobExecutionName: data.deploy_job_execution_name || undefined,
    deployJobStatus: data.deploy_job_status || undefined,
    deployJobStartedAt: data.deploy_job_started_at || undefined,
    deployJobCompletedAt: data.deploy_job_completed_at || undefined,
    deployJobError: data.deploy_job_error || undefined,
    deployJobOperation: data.deploy_job_operation || undefined,
    deployJobCurrentStep: data.deploy_job_current_step ?? undefined,
    deployJobTotalSteps: data.deploy_job_total_steps ?? undefined,
    deployJobStepDescription: data.deploy_job_step_description || undefined,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
    dnsConfigurations: dnsConfigurations.map((ip) => ({
      serviceName: ip.service_name,
      ip: ip.ip,
      sshRecordId: ip.ssh_record_id || undefined,
      routeRecordIds: ip.route_record_ids || undefined,
      routeRecordMap: ip.route_record_map || undefined,
      domainRoutes: ip.domain_routes || undefined,
    })),
  }
}

/**
 * Get installation by ID with dns_configurations
 */
export async function getInstallationById(installationId: string): Promise<Installation | null> {
  const result = await supabase.from('installations').select('*').eq('id', installationId).single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting installation:', result.error)
    return null
  }

  const data = result.data as InstallationRow | null
  if (!data) return null

  const ipResult = await supabase
    .from('dns_configurations')
    .select('*')
    .eq('installation_id', installationId)

  return mapToInstallation(data, (ipResult.data || []) as DnsConfigurationRow[])
}

/**
 * Get installation by installation key
 */
export async function getInstallationByKey(installationKey: string): Promise<Installation | null> {
  const result = await supabase
    .from('installations')
    .select('*')
    .eq('installation_key', installationKey)
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting installation by key:', result.error)
    return null
  }

  const data = result.data as InstallationRow | null
  if (!data) return null

  return getInstallationById(data.id)
}

/**
 * Get installation by secret key
 * Used for API authentication (e.g., Claude proxy)
 */
export async function getInstallationBySecretKey(secretKey: string): Promise<Installation | null> {
  const result = await supabase
    .from('installations')
    .select('*')
    .eq('secret_key', secretKey)
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting installation by secret key:', result.error)
    return null
  }

  const data = result.data as InstallationRow | null
  if (!data) return null

  return getInstallationById(data.id)
}

/**
 * Get all installations for an organization
 */
export async function getOrgInstallations(orgId: string): Promise<Installation[]> {
  const { data, error } = await supabase
    .from('installations')
    .select('*')
    .eq('org_id', orgId)
    .order('created_at', { ascending: false })

  if (error) {
    console.error('Error getting org installations:', error)
    return []
  }

  const installations = (data || []) as InstallationRow[]

  // Fetch dns_configurations for all installations in parallel
  return Promise.all(
    installations.map(async (inst) => {
      const ipResult = await supabase.from('dns_configurations').select('*').eq('installation_id', inst.id)
      return mapToInstallation(inst, (ipResult.data || []) as DnsConfigurationRow[])
    }),
  )
}

/**
 * Get only valid (non-expired) installations for an organization
 */
export async function getValidOrgInstallations(orgId: string): Promise<Installation[]> {
  const allInstallations = await getOrgInstallations(orgId)
  return allInstallations.filter(isInstallationValid)
}

/**
 * Cleanup expired installations for an organization
 * Deletes installations that:
 * 1. Are not deployment ready (domain not registered)
 * 2. Were created more than 15 minutes ago
 *
 * Also cleans up related records (domain_reservations, dns_configurations)
 * Returns the number of installations deleted
 */
export async function cleanupExpiredInstallations(orgId: string): Promise<number> {
  const allInstallations = await getOrgInstallations(orgId)
  const expiredInstallations = allInstallations.filter(inst => !isInstallationValid(inst))

  let deletedCount = 0

  for (const installation of expiredInstallations) {
    try {
      await supabase.from('domain_reservations').delete().eq('installation_id', installation.id)
      await supabase.from('dns_configurations').delete().eq('installation_id', installation.id)
      await supabase.from('installations').delete().eq('id', installation.id)

      deletedCount++
      console.log(`Cleaned up expired installation: ${installation.id} (${installation.name})`)
    } catch (error) {
      console.error(`Failed to cleanup installation ${installation.id}:`, error)
    }
  }

  return deletedCount
}

/**
 * Create a new installation within an organization
 */
export async function createInstallation(
  orgId: string,
  name: string,
  description: string | undefined,
  installationKey: string,
  subdomain?: string,
): Promise<Installation> {
  type InstallationInsert = Database['public']['Tables']['installations']['Insert']

  const insertData: InstallationInsert = {
    org_id: orgId,
    name: name,
    description: description,
    installation_key: installationKey,
    setup_completed: false,
    subdomain: subdomain || null,
  }

  const result = await supabase
    .from('installations')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert(insertData)
    .select()
    .single()

  if (result.error) {
    console.error('Error creating installation:', result.error)
    throw new Error(`Failed to create installation: ${result.error.message}`)
  }

  const data = result.data as InstallationRow

  return mapToInstallation(data)
}

/**
 * Update an installation
 */
export async function updateInstallation(
  installationId: string,
  updates: Partial<Installation>,
): Promise<Installation> {
  type InstallationUpdate = Database['public']['Tables']['installations']['Update']

  const updateData: InstallationUpdate = {}

  if (updates.secretKey !== undefined) updateData.secret_key = updates.secretKey || null
  if (updates.setupCompleted !== undefined)
    updateData.setup_completed = updates.setupCompleted
  if (updates.subdomain !== undefined) updateData.subdomain = updates.subdomain || null
  if (updates.reservedAt !== undefined) updateData.reserved_at = updates.reservedAt || null
  if (updates.deploymentReady !== undefined)
    updateData.deployment_ready = updates.deploymentReady || null
  if (updates.lastHealthCheck !== undefined)
    updateData.last_health_check = updates.lastHealthCheck || null
  if (updates.cloudProvider !== undefined)
    updateData.cloud_provider = updates.cloudProvider || null
  if (updates.cloudLocation !== undefined)
    updateData.cloud_location = updates.cloudLocation || null
  if (updates.deployJobExecutionName !== undefined)
    updateData.deploy_job_execution_name = updates.deployJobExecutionName || null
  if (updates.deployJobStatus !== undefined)
    updateData.deploy_job_status = updates.deployJobStatus || null
  if (updates.deployJobStartedAt !== undefined)
    updateData.deploy_job_started_at = updates.deployJobStartedAt || null
  if (updates.deployJobCompletedAt !== undefined)
    updateData.deploy_job_completed_at = updates.deployJobCompletedAt || null
  if (updates.deployJobError !== undefined)
    updateData.deploy_job_error = updates.deployJobError || null
  if (updates.deployJobOperation !== undefined)
    updateData.deploy_job_operation = updates.deployJobOperation || null
  if (updates.deployJobCurrentStep !== undefined)
    updateData.deploy_job_current_step = updates.deployJobCurrentStep ?? null
  if (updates.deployJobTotalSteps !== undefined)
    updateData.deploy_job_total_steps = updates.deployJobTotalSteps ?? null
  if (updates.deployJobStepDescription !== undefined)
    updateData.deploy_job_step_description = updates.deployJobStepDescription || null

  const { error } = await supabase
    .from('installations')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update(updateData)
    .eq('id', installationId)

  if (error) {
    throw new Error(`Failed to update installation: ${error.message}`)
  }

  const installation = await getInstallationById(installationId)
  if (!installation) {
    throw new Error('Installation not found after update')
  }
  return installation
}

/**
 * Atomically mark installation complete with optional secret key
 */
export async function markInstallationComplete(
  installationId: string,
  secretKey?: string,
): Promise<Installation> {
  return updateInstallation(installationId, {
    setupCompleted: true,
    secretKey,
  })
}

/**
 * Atomically update health check timestamp
 */
export async function updateHealthCheck(installationId: string): Promise<Installation> {
  return updateInstallation(installationId, {
    lastHealthCheck: new Date().toISOString(),
  })
}

/**
 * Delete an installation and its related dns_configurations
 */
export async function deleteInstallation(installationId: string): Promise<void> {
  const { error: ipError } = await supabase
    .from('dns_configurations')
    .delete()
    .eq('installation_id', installationId)

  if (ipError) {
    console.error('Error deleting dns_configurations:', ipError)
  }

  const { error } = await supabase.from('installations').delete().eq('id', installationId)

  if (error) {
    throw new Error(`Failed to delete installation: ${error.message}`)
  }
}

/**
 * Atomically mark deployment ready
 */
export async function markDeploymentReady(installationId: string, ready: boolean): Promise<void> {
  await updateInstallation(installationId, { deploymentReady: ready })
}

/**
 * Update root DNS info for an installation
 */
export async function updateInstallationRootDns(
  installationId: string,
  target: string,
  type: 'cname' | 'a',
  recordId: string,
): Promise<void> {
  type InstallationUpdate = Database['public']['Tables']['installations']['Update']
  const updateData: InstallationUpdate = {
    root_dns_target: target,
    root_dns_type: type,
    root_dns_record_id: recordId,
  }

  const { error } = await supabase
    .from('installations')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update(updateData)
    .eq('id', installationId)

  if (error) {
    throw new Error(`Failed to update root DNS: ${error.message}`)
  }
}

/**
 * Reset installation (clears subdomain, deployment status)
 */
export async function resetInstallation(installationId: string): Promise<void> {
  const { error } = await supabase
    .from('installations')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update({
      subdomain: null,
      reserved_at: null,
      secret_key: null,
      setup_completed: false,
      deployment_ready: false,
      last_health_check: null,
    })
    .eq('id', installationId)

  if (error) {
    throw new Error(`Failed to reset installation: ${error.message}`)
  }
}

// Installation validity window (15 minutes) - after this, incomplete installations expire
const INSTALLATION_VALIDITY_MS = 15 * 60 * 1000

/**
 * Check if an installation is still valid
 */
function isInstallationValid(installation: Installation): boolean {
  if (installation.deploymentReady) {
    return true
  }
  // Installations that have completed setup (paid) are always valid
  if (installation.setupCompleted) {
    return true
  }
  const createdAt = new Date(installation.createdAt).getTime()
  const now = Date.now()
  return (now - createdAt) < INSTALLATION_VALIDITY_MS
}
