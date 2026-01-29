/**
 * Storage Service using Supabase (PostgreSQL)
 *
 * Provides atomic operations using SQL transactions
 * No eventual consistency - ACID guarantees
 *
 * Updated to support multiple installations per user
 */

import type { Database } from './supabase-types'
import { supabase } from './supabase'

type UserRegistrationRow = Database['public']['Tables']['user_registrations']['Row']
type InstallationRow = Database['public']['Tables']['installations']['Row']
type IPRecordRow = Database['public']['Tables']['ip_records']['Row']
type DomainReservationRow = Database['public']['Tables']['domain_reservations']['Row']
type InstallationMemberRow = Database['public']['Tables']['installation_members']['Row']
type InstallationInvitationRow = Database['public']['Tables']['installation_invitations']['Row']

// Team & Member types
export type MemberRole = 'owner' | 'admin' | 'member' | 'viewer'
export type InvitationStatus = 'pending' | 'accepted' | 'rejected' | 'expired'

export interface InstallationMember {
  id: string
  installationId: string
  userId: string
  role: MemberRole
  addedBy: string | null
  addedAt: string
  createdAt: string
  updatedAt: string
  // Populated from join
  userEmail?: string
  userName?: string
  userProviders?: string[]
}

export interface InstallationInvitation {
  id: string
  installationId: string
  email: string
  role: Exclude<MemberRole, 'owner'> // owner can't be invited
  invitedBy: string
  status: InvitationStatus
  expiresAt: string
  createdAt: string
  updatedAt: string
  // Populated from join
  inviterName?: string
  installationName?: string
}

export interface IPRecord {
  domainRequestName: string
  ip: string
  configuredAt: string
  sshRecordId?: string | null
  routeRecordIds?: string[] // Kept for backward compatibility
  routeRecordMap?: Record<string, string> // domain -> CNAME record ID
  domainRoutes?: Array<{ domain: string }> // List of domains
}

export interface Installation {
  id: string
  userId: string
  name?: string
  description?: string
  installationKey: string
  secretKey?: string
  hasCompletedInstallation: boolean
  subdomain?: string
  reservedAt?: string
  ipRecords?: IPRecord[]
  deploymentReady?: boolean
  lastHealthCheck?: string
  cloudProvider?: 'aws' | 'gcp' | 'azure'
  cloudLocation?: string
  createdAt: string
  updatedAt: string
}

export interface UserRegistration {
  userId: string
  email: string
  name: string
  providers: ('github' | 'google' | 'azure-ad')[]
  registeredAt: string
  createdAt: string
  updatedAt: string
}

export interface DomainReservation {
  subdomain: string
  installationId: string
  userId: string
  reservedAt: string
  status: 'reserved' | 'active' | 'cancelled'
  userEmail: string
  userName: string
}

/**
 * User Management Functions
 */

/**
 * Get user by userId
 */
export async function getUserById(userId: string): Promise<UserRegistration | null> {
  const result = await supabase
    .from('user_registrations')
    .select('*')
    .eq('user_id', userId)
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting user:', result.error)
    return null
  }

  const data = result.data as UserRegistrationRow | null
  if (!data) return null

  return {
    userId: data.user_id,
    email: data.email,
    name: data.name,
    providers: data.providers || [],
    registeredAt: data.registered_at,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Get user by email
 */
export async function getUserByEmail(email: string): Promise<UserRegistration | null> {
  const result = await supabase
    .from('user_registrations')
    .select('*')
    .eq('email', email.toLowerCase())
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting user:', result.error)
    return null
  }

  const data = result.data as UserRegistrationRow | null
  if (!data) return null

  return {
    userId: data.user_id,
    email: data.email,
    name: data.name,
    providers: data.providers || [],
    registeredAt: data.registered_at,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Create or update user registration
 */
export async function saveUserRegistration(registration: UserRegistration): Promise<void> {
  type UserRegistrationInsert = Database['public']['Tables']['user_registrations']['Insert']
  type UserRegistrationUpdate = Database['public']['Tables']['user_registrations']['Update']

  const insertData: UserRegistrationInsert = {
    user_id: registration.userId,
    email: registration.email.toLowerCase(),
    name: registration.name,
    providers: registration.providers,
    registered_at: registration.registeredAt,
  }

  // Try to insert first
  const { error: insertError } = await supabase
    .from('user_registrations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert(insertData)

  // If user already exists (unique constraint violation), update instead
  if (insertError && insertError.code === '23505') {
    const updateData: UserRegistrationUpdate = {
      name: registration.name,
      providers: registration.providers,
    }

    const { error: updateError } = await supabase
      .from('user_registrations')
      // @ts-expect-error - Supabase client with placeholder values has type issues during build
      .update(updateData)
      .eq('user_id', registration.userId)

    if (updateError) {
      console.error('Error updating user registration:', updateError)
      throw new Error(`Failed to update user registration: ${updateError.message}`)
    }
  } else if (insertError) {
    console.error('Error saving user registration:', insertError)
    throw new Error(`Failed to save user registration: ${insertError.message}`)
  }
}

/**
 * Installation Management Functions
 */

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
 * Get all installations for a user
 */
export async function getUserInstallations(userId: string): Promise<Installation[]> {
  const result = await supabase
    .from('installations')
    .select('*')
    .eq('user_id', userId)
    .order('created_at', { ascending: false })

  if (result.error) {
    console.error('Error getting user installations:', result.error)
    return []
  }

  const installations = (result.data || []) as InstallationRow[]

  // Fetch IP records for all installations in parallel
  const installationsWithIpRecords = await Promise.all(
    installations.map(async (inst) => {
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
 * IP Record Management
 */

/**
 * Atomically add or update IP record
 */
export async function addOrUpdateIpRecord(
  installationId: string,
  ipRecord: IPRecord,
): Promise<number> {
  const { error } = await supabase
    .from('ip_records')
    // @ts-expect-error - Supabase placeholder client type inference issue
    .upsert(
      {
        installation_id: installationId,
        domain_request_name: ipRecord.domainRequestName,
        ip: ipRecord.ip,
        configured_at: ipRecord.configuredAt,
        ssh_record_id: ipRecord.sshRecordId || null,
        route_record_ids: ipRecord.routeRecordIds || [],
        route_record_map: ipRecord.routeRecordMap || {},
        domain_routes: ipRecord.domainRoutes || [],
      },
      {
        onConflict: 'installation_id,domain_request_name',
      },
    )

  if (error) {
    throw new Error(`Failed to add IP record: ${error.message}`)
  }

  // Get total count
  const { count } = await supabase
    .from('ip_records')
    .select('*', { count: 'exact', head: true })
    .eq('installation_id', installationId)

  return count || 0
}

/**
 * Remove a specific IP record for a domain request
 */
export async function removeIpRecord(
  installationId: string,
  domainRequestName: string,
): Promise<void> {
  const { error } = await supabase
    .from('ip_records')
    .delete()
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)

  if (error) {
    console.error('Error removing IP record:', error)
    throw new Error(`Failed to remove IP record: ${error.message}`)
  }
}

/**
 * Delete all IP records for an installation and return their DNS record IDs
 */
export async function deleteIpRecords(installationId: string): Promise<string[]> {
  // Get all IP records to extract DNS record IDs
  const { data: ipData } = await supabase
    .from('ip_records')
    .select('*')
    .eq('installation_id', installationId)

  const dnsRecordIds: string[] = []

  if (ipData) {
    for (const record of ipData as IPRecordRow[]) {
      // Collect SSH record ID
      if (record.ssh_record_id) {
        dnsRecordIds.push(record.ssh_record_id)
      }
      // Collect route record IDs
      if (record.route_record_ids && Array.isArray(record.route_record_ids)) {
        dnsRecordIds.push(...record.route_record_ids)
      }
    }
  }

  // Delete all IP records
  const { error } = await supabase.from('ip_records').delete().eq('installation_id', installationId)

  if (error) {
    console.error('Error deleting IP records:', error)
  }

  return dnsRecordIds
}

/**
 * Domain Reservation Management
 */

/**
 * Reserved subdomains that cannot be used
 */
const RESERVED_KEYWORDS = [
  'www',
  'api',
  'admin',
  'mail',
  'smtp',
  'ftp',
  'ssh',
  'vpn',
  'dev',
  'staging',
  'prod',
  'production',
  'test',
  'demo',
  'app',
  'portal',
  'dashboard',
  'console',
  'docs',
  'blog',
  'status',
  'support',
  'help',
  'cdn',
  'static',
  'assets',
]

/**
 * Validate subdomain format
 * Returns { valid: true } or { valid: false, reason: string }
 */
export function validateSubdomain(subdomain: string): { valid: boolean; reason?: 'reserved' | 'invalid' } {
  // Check length: 3-63 characters
  if (subdomain.length < 3 || subdomain.length > 63) {
    return { valid: false, reason: 'invalid' }
  }

  // Check format: alphanumeric + hyphens, can't start/end with hyphen
  const subdomainRegex = /^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/i
  if (!subdomainRegex.test(subdomain)) {
    return { valid: false, reason: 'invalid' }
  }

  // Check reserved keywords
  if (RESERVED_KEYWORDS.includes(subdomain.toLowerCase())) {
    return { valid: false, reason: 'reserved' }
  }

  return { valid: true }
}

// 15 minutes in milliseconds
export const DOMAIN_RESERVATION_EXPIRY_MS = 15 * 60 * 1000

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

/**
 * Check if subdomain is available
 * A subdomain is considered available if:
 * 1. It doesn't exist in domain_reservations, OR
 * 2. It exists but the associated installation hasn't completed (deployment_ready=false)
 *    AND the reservation is older than 15 minutes
 */
export async function isSubdomainAvailable(subdomain: string): Promise<boolean> {
  // First validate the subdomain format
  const validation = validateSubdomain(subdomain)
  if (!validation.valid) {
    return false
  }

  const subdomainLower = subdomain.toLowerCase()

  // Check if subdomain exists in domain_reservations
  const reservationResult = await supabase
    .from('domain_reservations')
    .select('subdomain, installation_id, reserved_at')
    .eq('subdomain', subdomainLower)
    .single()

  // If no reservation exists, subdomain is available
  if (!reservationResult.data) {
    return true
  }

  // Reservation exists - check if it's expired
  // @ts-expect-error - Supabase client with placeholder values has type issues during build
  const reservation = reservationResult.data
  const reservedAt = new Date(reservation.reserved_at).getTime()
  const now = Date.now()
  const isExpired = (now - reservedAt) > DOMAIN_RESERVATION_EXPIRY_MS

  if (!isExpired) {
    // Reservation is still fresh, subdomain is not available
    return false
  }

  // Reservation is expired - check if installation has completed
  const installationResult = await supabase
    .from('installations')
    .select('deployment_ready')
    .eq('id', reservation.installation_id)
    .single()

  // If installation doesn't exist or hasn't completed deployment, subdomain is available
  // (the expired reservation will be cleaned up when someone reserves it)
  // @ts-expect-error - Supabase client with placeholder values has type issues during build
  if (!installationResult.data || !installationResult.data.deployment_ready) {
    return true
  }

  // Installation is complete, subdomain is not available
  return false
}

/**
 * Atomically reserve a subdomain for an installation
 */
export async function reserveSubdomain(
  subdomain: string,
  installationId: string,
  userId: string,
  userEmail: string,
  userName: string,
): Promise<DomainReservation> {
  const subdomainLower = subdomain.toLowerCase()
  const reservedAt = new Date().toISOString()

  // Insert domain reservation (will fail if subdomain already exists due to PRIMARY KEY)
  const result = await supabase
    .from('domain_reservations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert({
      subdomain: subdomainLower,
      installation_id: installationId,
      user_id: userId,
      user_email: userEmail.toLowerCase(),
      user_name: userName,
      reserved_at: reservedAt,
      status: 'reserved',
    })
    .select()
    .single()

  if (result.error) {
    if (result.error.code === '23505') {
      // Unique constraint violation
      throw new Error('Subdomain is already reserved')
    }
    throw new Error(`Failed to reserve subdomain: ${result.error.message}`)
  }

  const data = result.data as DomainReservationRow

  // Atomically update installation with subdomain and mark as ready for deployment
  await supabase
    .from('installations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update({
      subdomain: subdomainLower,
      reserved_at: reservedAt,
      deployment_ready: true,
    })
    .eq('id', installationId)

  return {
    subdomain: data.subdomain,
    installationId: data.installation_id,
    userId: data.user_id,
    reservedAt: data.reserved_at,
    status: data.status,
    userEmail: data.user_email,
    userName: data.user_name,
  }
}

/**
 * Get domain reservation by subdomain
 */
export async function getDomainReservation(subdomain: string): Promise<DomainReservation | null> {
  const result = await supabase
    .from('domain_reservations')
    .select('*')
    .eq('subdomain', subdomain.toLowerCase())
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting domain reservation:', result.error)
    return null
  }

  const data = result.data as DomainReservationRow | null
  if (!data) return null

  return {
    subdomain: data.subdomain,
    installationId: data.installation_id,
    userId: data.user_id,
    reservedAt: data.reserved_at,
    status: data.status,
    userEmail: data.user_email,
    userName: data.user_name,
  }
}

/**
 * Check if an installation's domain reservation has expired and been claimed by someone else
 * Returns information about the domain status
 */
export async function checkInstallationDomainStatus(
  installationId: string,
  subdomain: string,
): Promise<{
  isExpired: boolean
  isClaimedByOther: boolean
  claimedByEmail?: string
  claimedByName?: string
}> {
  const subdomainLower = subdomain.toLowerCase()

  // Get the current reservation for this subdomain
  const reservationResult = await supabase
    .from('domain_reservations')
    .select('subdomain, installation_id, user_id, user_email, user_name, reserved_at')
    .eq('subdomain', subdomainLower)
    .single()

  // If no reservation exists, domain was never reserved or got cleaned up
  if (!reservationResult.data) {
    return { isExpired: true, isClaimedByOther: false }
  }

  // @ts-expect-error - Supabase client with placeholder values has type issues during build
  const reservation = reservationResult.data

  // If the reservation belongs to this installation, it's not claimed by another
  if (reservation.installation_id === installationId) {
    // Check if it's expired (>15 min since reservation)
    const reservedAt = new Date(reservation.reserved_at).getTime()
    const now = Date.now()
    const isExpired = now - reservedAt > DOMAIN_RESERVATION_EXPIRY_MS

    return { isExpired, isClaimedByOther: false }
  }

  // Domain is reserved by a different installation
  // Check if that installation has completed deployment
  const installationResult = await supabase
    .from('installations')
    .select('deployment_ready')
    .eq('id', reservation.installation_id)
    .single()

  // If the other installation has completed deployment, the domain is permanently claimed
  // @ts-expect-error - Supabase client with placeholder values has type issues during build
  if (installationResult.data?.deployment_ready) {
    return {
      isExpired: true,
      isClaimedByOther: true,
      claimedByEmail: reservation.user_email,
      claimedByName: reservation.user_name,
    }
  }

  // The other reservation might be expired too, check its age
  const reservedAt = new Date(reservation.reserved_at).getTime()
  const now = Date.now()
  const otherIsExpired = now - reservedAt > DOMAIN_RESERVATION_EXPIRY_MS

  if (otherIsExpired) {
    // The other reservation is expired and not deployed, domain is available
    return { isExpired: true, isClaimedByOther: false }
  }

  // The other reservation is still active (within 15 min window)
  return {
    isExpired: true,
    isClaimedByOther: true,
    claimedByEmail: reservation.user_email,
    claimedByName: reservation.user_name,
  }
}

/**
 * Re-reserve a new subdomain for an existing installation
 * This is used when the original domain has expired and been claimed by someone else
 */
export async function reReserveSubdomain(
  installationId: string,
  newSubdomain: string,
  userId: string,
  userEmail: string,
  userName: string,
): Promise<DomainReservation> {
  const subdomainLower = newSubdomain.toLowerCase()
  const reservedAt = new Date().toISOString()

  // First delete the old domain reservation for this installation (if any)
  await supabase.from('domain_reservations').delete().eq('installation_id', installationId)

  // Insert new domain reservation
  const result = await supabase
    .from('domain_reservations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert({
      subdomain: subdomainLower,
      installation_id: installationId,
      user_id: userId,
      user_email: userEmail.toLowerCase(),
      user_name: userName,
      reserved_at: reservedAt,
      status: 'reserved',
    })
    .select()
    .single()

  if (result.error) {
    if (result.error.code === '23505') {
      throw new Error('Subdomain is already reserved')
    }
    throw new Error(`Failed to reserve subdomain: ${result.error.message}`)
  }

  const data = result.data as DomainReservationRow

  // Update installation with new subdomain
  await supabase
    .from('installations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update({
      subdomain: subdomainLower,
      reserved_at: reservedAt,
    })
    .eq('id', installationId)

  return {
    subdomain: data.subdomain,
    installationId: data.installation_id,
    userId: data.user_id,
    reservedAt: data.reserved_at,
    status: data.status,
    userEmail: data.user_email,
    userName: data.user_name,
  }
}

/**
 * Delete domain reservation for an installation
 */
export async function deleteDomainReservation(installationId: string): Promise<void> {
  const { error } = await supabase
    .from('domain_reservations')
    .delete()
    .eq('installation_id', installationId)

  if (error) {
    console.error('Error deleting domain reservation:', error)
    throw new Error(`Failed to delete domain reservation: ${error.message}`)
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

// ============================================================================
// TEAM MEMBER MANAGEMENT
// ============================================================================

/**
 * Get member's role for an installation
 * Returns null if not a member
 */
export async function getMemberRole(
  installationId: string,
  userId: string
): Promise<MemberRole | null> {
  const { data, error } = await supabase
    .from('installation_members')
    .select('role')
    .eq('installation_id', installationId)
    .eq('user_id', userId)
    .single()

  if (error) {
    if (error.code === 'PGRST116') return null // No rows returned
    console.error('Error getting member role:', error)
    return null
  }

  return data.role as MemberRole
}

/**
 * Check if user has permission to access installation
 */
export async function canAccessInstallation(
  installationId: string,
  userId: string
): Promise<boolean> {
  const role = await getMemberRole(installationId, userId)
  return role !== null
}

/**
 * Check if user has admin or owner permission
 */
export async function canManageInstallation(
  installationId: string,
  userId: string
): Promise<boolean> {
  const role = await getMemberRole(installationId, userId)
  return role === 'owner' || role === 'admin'
}

/**
 * Get all members for an installation with user details
 */
export async function getInstallationMembers(
  installationId: string
): Promise<InstallationMember[]> {
  // First get the members
  const { data: membersData, error: membersError } = await supabase
    .from('installation_members')
    .select('*')
    .eq('installation_id', installationId)
    .order('added_at', { ascending: true })

  if (membersError) {
    console.error('Error getting installation members:', membersError)
    return []
  }

  if (!membersData || membersData.length === 0) {
    return []
  }

  // Get unique user IDs
  const userIds = [...new Set(membersData.map((m: any) => m.user_id))]

  // Fetch user details
  const { data: usersData, error: usersError } = await supabase
    .from('user_registrations')
    .select('user_id, email, name, providers')
    .in('user_id', userIds)

  if (usersError) {
    console.error('Error getting user details:', usersError)
  }

  // Create a map of user details
  const userMap = new Map(
    (usersData || []).map((u: any) => [u.user_id, u])
  )

  return membersData.map((row: any) => {
    const user = userMap.get(row.user_id) || {}
    return {
      id: row.id,
      installationId: row.installation_id,
      userId: row.user_id,
      role: row.role,
      addedBy: row.added_by,
      addedAt: row.added_at,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
      userEmail: user.email || 'Unknown',
      userName: user.name || 'Unknown User',
      userProviders: user.providers || [],
    }
  })
}

/**
 * Add a member to an installation
 */
export async function addInstallationMember(
  installationId: string,
  userId: string,
  role: MemberRole,
  addedBy: string
): Promise<InstallationMember> {
  const { data, error } = await supabase
    .from('installation_members')
    .insert({
      installation_id: installationId,
      user_id: userId,
      role,
      added_by: addedBy,
    })
    .select()
    .single()

  if (error) {
    throw new Error(`Failed to add member: ${error.message}`)
  }

  return {
    id: data.id,
    installationId: data.installation_id,
    userId: data.user_id,
    role: data.role,
    addedBy: data.added_by,
    addedAt: data.added_at,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Update member role
 */
export async function updateMemberRole(
  memberId: string,
  newRole: MemberRole
): Promise<void> {
  const { error } = await supabase
    .from('installation_members')
    .update({ role: newRole })
    .eq('id', memberId)

  if (error) {
    throw new Error(`Failed to update member role: ${error.message}`)
  }
}

/**
 * Remove a member from installation
 */
export async function removeInstallationMember(memberId: string): Promise<void> {
  const { error } = await supabase
    .from('installation_members')
    .delete()
    .eq('id', memberId)

  if (error) {
    throw new Error(`Failed to remove member: ${error.message}`)
  }
}

// ============================================================================
// INVITATION MANAGEMENT
// ============================================================================

/**
 * Create invitation for email
 */
export async function createInvitation(
  installationId: string,
  email: string,
  role: Exclude<MemberRole, 'owner'>,
  invitedBy: string
): Promise<InstallationInvitation> {
  const emailLower = email.toLowerCase()

  const { data, error} = await supabase
    .from('installation_invitations')
    .insert({
      installation_id: installationId,
      email: emailLower,
      role,
      invited_by: invitedBy,
      status: 'pending' as const,
      expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    })
    .select()
    .single()

  if (error) {
    if (error.code === '23505') {
      throw new Error('User already has a pending invitation')
    }
    throw new Error(`Failed to create invitation: ${error.message}`)
  }

  return {
    id: data.id,
    installationId: data.installation_id,
    email: data.email,
    role: data.role as Exclude<MemberRole, 'owner'>,
    invitedBy: data.invited_by,
    status: data.status,
    expiresAt: data.expires_at,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Get pending invitations for an installation
 */
export async function getInstallationInvitations(
  installationId: string
): Promise<InstallationInvitation[]> {
  const { data, error } = await supabase
    .from('installation_invitations')
    .select(`
      *,
      user_registrations!installation_invitations_invited_by_fkey(name),
      installations!inner(name)
    `)
    .eq('installation_id', installationId)
    .eq('status', 'pending')
    .order('created_at', { ascending: false })

  if (error) {
    console.error('Error getting invitations:', error)
    return []
  }

  return data.map((row: any) => ({
    id: row.id,
    installationId: row.installation_id,
    email: row.email,
    role: row.role,
    invitedBy: row.invited_by,
    status: row.status,
    expiresAt: row.expires_at,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    inviterName: row.user_registrations?.name,
    installationName: row.installations?.name,
  }))
}

/**
 * Get pending invitations for a user's email
 */
export async function getUserPendingInvitations(
  email: string
): Promise<InstallationInvitation[]> {
  const emailLower = email.toLowerCase()
  const now = new Date().toISOString()

  const { data, error } = await supabase
    .from('installation_invitations')
    .select(`
      *,
      user_registrations!installation_invitations_invited_by_fkey(name),
      installations!inner(name)
    `)
    .eq('email', emailLower)
    .eq('status', 'pending')
    .gt('expires_at', now)
    .order('created_at', { ascending: false })

  if (error) {
    console.error('Error getting user invitations:', error)
    return []
  }

  return data.map((row: any) => ({
    id: row.id,
    installationId: row.installation_id,
    email: row.email,
    role: row.role,
    invitedBy: row.invited_by,
    status: row.status,
    expiresAt: row.expires_at,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    inviterName: row.user_registrations?.name,
    installationName: row.installations?.name,
  }))
}

/**
 * Accept an invitation
 */
export async function acceptInvitation(
  invitationId: string,
  userId: string
): Promise<InstallationMember> {
  // Get invitation details
  const { data: invitation, error: invError } = await supabase
    .from('installation_invitations')
    .select('*')
    .eq('id', invitationId)
    .single()

  if (invError || !invitation) {
    throw new Error('Invitation not found')
  }

  // Check if expired
  if (new Date(invitation.expires_at) < new Date()) {
    throw new Error('Invitation has expired')
  }

  // Add member
  const member = await addInstallationMember(
    invitation.installation_id,
    userId,
    invitation.role,
    invitation.invited_by
  )

  // Mark invitation as accepted
  await supabase
    .from('installation_invitations')
    .update({ status: 'accepted' })
    .eq('id', invitationId)

  return member
}

/**
 * Reject an invitation
 */
export async function rejectInvitation(invitationId: string): Promise<void> {
  const { error } = await supabase
    .from('installation_invitations')
    .update({ status: 'rejected' })
    .eq('id', invitationId)

  if (error) {
    throw new Error(`Failed to reject invitation: ${error.message}`)
  }
}

/**
 * Delete/cancel an invitation
 */
export async function deleteInvitation(invitationId: string): Promise<void> {
  const { error } = await supabase
    .from('installation_invitations')
    .delete()
    .eq('id', invitationId)

  if (error) {
    throw new Error(`Failed to delete invitation: ${error.message}`)
  }
}

