/**
 * Installation Management Functions
 */

import type { Database } from '../supabase-types'
import { supabase } from '../supabase'
import type { Installation, InstallationRow, IPRecordRow } from './types'

/**
 * Get installation by ID with IP records
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

  // Fetch IP records
  const ipResult = await supabase
    .from('ip_records')
    .select('*')
    .eq('installation_id', installationId)

  return {
    id: data.id,
    userId: data.user_id,
    name: data.name || undefined,
    description: data.description || undefined,
    installationKey: data.installation_key,
    secretKey: data.secret_key || undefined,
    hasCompletedInstallation: data.has_completed_installation,
    subdomain: data.subdomain || undefined,
    reservedAt: data.reserved_at || undefined,
    deploymentReady: data.deployment_ready || undefined,
    lastHealthCheck: data.last_health_check || undefined,
    cloudProvider: data.cloud_provider || undefined,
    cloudLocation: data.cloud_location || undefined,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
    ipRecords:
      ((ipResult.data || []) as IPRecordRow[]).map((ip) => ({
        domainRequestName: ip.domain_request_name,
        ip: ip.ip,
        configuredAt: ip.configured_at,
        sshRecordId: ip.ssh_record_id || undefined,
        routeRecordIds: ip.route_record_ids || undefined,
        routeRecordMap: ip.route_record_map || undefined,
        domainRoutes: ip.domain_routes || undefined,
      })) || [],
  }
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
 * Get all installations for a user (owned + member of)
 */
export async function getUserInstallations(userId: string): Promise<Installation[]> {
  // Get installations owned by the user
  const ownedResult = await supabase
    .from('installations')
    .select('*')
    .eq('user_id', userId)
    .order('created_at', { ascending: false })

  if (ownedResult.error) {
    console.error('Error getting owned installations:', ownedResult.error)
  }

  const ownedInstallations = (ownedResult.data || []) as InstallationRow[]

  // Get installation IDs where the user is a member
  const memberResult = await supabase
    .from('installation_members')
    .select('installation_id')
    .eq('user_id', userId)

  if (memberResult.error) {
    console.error('Error getting member installations:', memberResult.error)
  }

  const memberInstallationIds = (memberResult.data || []).map((m: any) => m.installation_id)

  // Get owned installation IDs
  const ownedIds = new Set(ownedInstallations.map((i) => i.id))

  // Filter to only IDs not already owned
  const additionalIds = memberInstallationIds.filter((id: string) => !ownedIds.has(id))

  // Fetch additional installations where user is a member but not owner
  let memberInstallations: InstallationRow[] = []
  if (additionalIds.length > 0) {
    const additionalResult = await supabase
      .from('installations')
      .select('*')
      .in('id', additionalIds)
      .order('created_at', { ascending: false })

    if (additionalResult.error) {
      console.error('Error getting additional installations:', additionalResult.error)
    } else {
      memberInstallations = (additionalResult.data || []) as InstallationRow[]
    }
  }

  // Combine owned and member installations
  const allInstallations = [...ownedInstallations, ...memberInstallations]

  // Fetch IP records for all installations in parallel
  const installationsWithIpRecords = await Promise.all(
    allInstallations.map(async (inst) => {
      const ipResult = await supabase.from('ip_records').select('*').eq('installation_id', inst.id)

      return {
        id: inst.id,
        userId: inst.user_id,
        name: inst.name || undefined,
        description: inst.description || undefined,
        installationKey: inst.installation_key,
        secretKey: inst.secret_key || undefined,
        hasCompletedInstallation: inst.has_completed_installation,
        subdomain: inst.subdomain || undefined,
        reservedAt: inst.reserved_at || undefined,
        deploymentReady: inst.deployment_ready || undefined,
        lastHealthCheck: inst.last_health_check || undefined,
        cloudProvider: inst.cloud_provider || undefined,
        cloudLocation: inst.cloud_location || undefined,
        createdAt: inst.created_at,
        updatedAt: inst.updated_at,
        ipRecords:
          ((ipResult.data || []) as IPRecordRow[]).map((ip) => ({
            domainRequestName: ip.domain_request_name,
            ip: ip.ip,
            configuredAt: ip.configured_at,
            sshRecordId: ip.ssh_record_id || undefined,
            routeRecordIds: ip.route_record_ids || undefined,
            routeRecordMap: ip.route_record_map || undefined,
            domainRoutes: ip.domain_routes || undefined,
          })) || [],
      }
    }),
  )

  return installationsWithIpRecords
}

/**
 * Get only valid (non-expired) installations for a user
 */
export async function getValidUserInstallations(userId: string): Promise<Installation[]> {
  const allInstallations = await getUserInstallations(userId)
  return allInstallations.filter(isInstallationValid)
}

/**
 * Cleanup expired installations for a user
 * Deletes installations that:
 * 1. Are not deployment ready (domain not registered)
 * 2. Were created more than 15 minutes ago
 *
 * Also cleans up related records (domain_reservations, ip_records)
 * Returns the number of installations deleted
 */
export async function cleanupExpiredInstallations(userId: string): Promise<number> {
  const allInstallations = await getUserInstallations(userId)
  const expiredInstallations = allInstallations.filter(inst => !isInstallationValid(inst))

  let deletedCount = 0

  for (const installation of expiredInstallations) {
    try {
      // Delete domain reservation
      await supabase
        .from('domain_reservations')
        .delete()
        .eq('installation_id', installation.id)

      // Delete IP records
      await supabase
        .from('ip_records')
        .delete()
        .eq('installation_id', installation.id)

      // Delete the installation itself
      await supabase
        .from('installations')
        .delete()
        .eq('id', installation.id)

      deletedCount++
      console.log(`Cleaned up expired installation: ${installation.id} (${installation.name})`)
    } catch (error) {
      console.error(`Failed to cleanup installation ${installation.id}:`, error)
    }
  }

  return deletedCount
}

/**
 * Create a new installation
 */
export async function createInstallation(
  userId: string,
  name: string,
  description: string | undefined,
  installationKey: string,
  subdomain?: string,
): Promise<Installation> {
  type InstallationInsert = Database['public']['Tables']['installations']['Insert']

  const insertData: InstallationInsert = {
    user_id: userId,
    name: name,
    description: description,
    installation_key: installationKey,
    has_completed_installation: false,
    subdomain: subdomain || null,
  }

  const result = await supabase
    .from('installations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert(insertData)
    .select()
    .single()

  if (result.error) {
    console.error('Error creating installation:', result.error)
    throw new Error(`Failed to create installation: ${result.error.message}`)
  }

  const data = result.data as InstallationRow

  // Add the creator as owner in installation_members
  const memberResult = await (supabase as any)
    .from('installation_members')
    .insert({
      installation_id: data.id,
      user_id: userId,
      role: 'owner',
      added_by: userId,
    })

  if (memberResult.error) {
    console.error('Error adding owner to installation_members:', memberResult.error)
    // Don't throw - the installation is created, member just failed to be added
  }

  return {
    id: data.id,
    userId: data.user_id,
    name: data.name || undefined,
    description: data.description || undefined,
    installationKey: data.installation_key,
    secretKey: data.secret_key || undefined,
    hasCompletedInstallation: data.has_completed_installation,
    subdomain: data.subdomain || undefined,
    reservedAt: data.reserved_at || undefined,
    deploymentReady: data.deployment_ready || undefined,
    lastHealthCheck: data.last_health_check || undefined,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
    ipRecords: [],
  }
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
  if (updates.hasCompletedInstallation !== undefined)
    updateData.has_completed_installation = updates.hasCompletedInstallation
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

  const { error } = await supabase
    .from('installations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
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
    hasCompletedInstallation: true,
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
 * Delete an installation and its related IP records
 */
export async function deleteInstallation(installationId: string): Promise<void> {
  // First delete related IP records
  const { error: ipError } = await supabase
    .from('ip_records')
    .delete()
    .eq('installation_id', installationId)

  if (ipError) {
    console.error('Error deleting IP records:', ipError)
    // Continue anyway - we still want to try to delete the installation
  }

  // Delete the installation
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
 * Supports both CNAME (for load balancers) and A records (for direct IPs)
 */
export async function updateInstallationRootDns(
  installationId: string,
  target: string,
  type: 'cname' | 'a',
  recordId: string,
): Promise<void> {
  const updateData = {
    root_dns_target: target,
    root_dns_type: type,
    root_dns_record_id: recordId,
    updated_at: new Date().toISOString(),
  }

  const { error } = await supabase
    .from('installations')
    // @ts-expect-error - These columns exist in DB but not in generated types
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
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update({
      subdomain: null,
      reserved_at: null,
      secret_key: null,
      has_completed_installation: false,
      deployment_ready: false,
      last_health_check: null,
    })
    .eq('id', installationId)

  if (error) {
    throw new Error(`Failed to reset installation: ${error.message}`)
  }
}

// Installation validity window (15 minutes) - after this, incomplete installations expire
export const INSTALLATION_VALIDITY_MS = 15 * 60 * 1000

/**
 * Check if an installation is still valid
 * An installation is valid if:
 * 1. It has completed deployment (deploymentReady === true), OR
 * 2. It was created within the last 15 minutes
 */
export function isInstallationValid(installation: Installation): boolean {
  // If domain is registered and deployment is ready, it's always valid
  if (installation.deploymentReady) {
    return true
  }
  // Otherwise, check if within 15-minute validity window
  const createdAt = new Date(installation.createdAt).getTime()
  const now = Date.now()
  return (now - createdAt) < INSTALLATION_VALIDITY_MS
}

/**
 * Get remaining validity time in milliseconds
 * Returns 0 if expired, or remaining time if still valid
 */
export function getInstallationRemainingTime(installation: Installation): number {
  if (installation.deploymentReady) {
    return Infinity // Always valid
  }
  const createdAt = new Date(installation.createdAt).getTime()
  const expiresAt = createdAt + INSTALLATION_VALIDITY_MS
  const remaining = expiresAt - Date.now()
  return remaining > 0 ? remaining : 0
}
