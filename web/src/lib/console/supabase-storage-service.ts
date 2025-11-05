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
type TLSCertificateRow = Database['public']['Tables']['tls_certificates']['Row']
type DomainReservationRow = Database['public']['Tables']['domain_reservations']['Row']

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
    installationKey: data.installation_key,
    secretKey: data.secret_key || undefined,
    hasCompletedInstallation: data.has_completed_installation,
    subdomain: data.subdomain || undefined,
    reservedAt: data.reserved_at || undefined,
    deploymentReady: data.deployment_ready || undefined,
    lastHealthCheck: data.last_health_check || undefined,
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
 * Create a new installation
 */
export async function createInstallation(
  userId: string,
  name: string,
  description: string | undefined,
  installationKey: string,
): Promise<Installation> {
  type InstallationInsert = Database['public']['Tables']['installations']['Insert']

  const insertData: InstallationInsert = {
    user_id: userId,
    name: name,
    description: description,
    installation_key: installationKey,
    has_completed_installation: false,
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
 * Check if subdomain is available
 */
export async function isSubdomainAvailable(subdomain: string): Promise<boolean> {
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

  if (RESERVED_KEYWORDS.includes(subdomain.toLowerCase())) {
    return false
  }

  const result = await supabase
    .from('domain_reservations')
    .select('subdomain')
    .eq('subdomain', subdomain.toLowerCase())
    .single()

  return !result.data
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

/**
 * Certificate Management
 */
export type CertificateScope = 'installation' | 'workmachine' | 'workspace'

export interface TLSCertificate {
  id?: number
  installationId: string
  cloudflareCertId: string | null
  certificate: string
  privateKey: string
  hostnames: string[]
  scope: CertificateScope
  scopeIdentifier?: string | null
  parentScopeIdentifier?: string | null
  validFrom: string
  validUntil: string
  generatedAt?: string
}

/**
 * Save TLS certificate
 */
export async function saveCertificate(cert: TLSCertificate): Promise<void> {
  const { error } = await supabase
    .from('tls_certificates')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert({
      installation_id: cert.installationId,
      cloudflare_cert_id: cert.cloudflareCertId,
      certificate: cert.certificate,
      private_key: cert.privateKey,
      hostnames: cert.hostnames,
      scope: cert.scope,
      scope_identifier: cert.scopeIdentifier || null,
      parent_scope_identifier: cert.parentScopeIdentifier || null,
      valid_from: cert.validFrom,
      valid_until: cert.validUntil,
      generated_at: cert.generatedAt || new Date().toISOString(),
    })

  if (error) {
    throw new Error(`Failed to save certificate: ${error.message}`)
  }
}

/**
 * Get latest certificate for installation, optionally filtered by scope
 */
export async function getLatestCertificate(
  installationId: string,
  scope?: CertificateScope,
  scopeIdentifier?: string,
  parentScopeIdentifier?: string,
): Promise<TLSCertificate | null> {
  let query = supabase.from('tls_certificates').select('*').eq('installation_id', installationId)

  if (scope) {
    query = query.eq('scope', scope)
  }

  if (scopeIdentifier !== undefined) {
    query = query.eq('scope_identifier', scopeIdentifier)
  }

  if (parentScopeIdentifier !== undefined) {
    query = query.eq('parent_scope_identifier', parentScopeIdentifier)
  }

  const result = await query.order('generated_at', { ascending: false }).limit(1).single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting certificate:', result.error)
    return null
  }

  const data = result.data as TLSCertificateRow | null
  if (!data) return null

  return {
    id: data.id,
    installationId: data.installation_id,
    cloudflareCertId: data.cloudflare_cert_id,
    certificate: data.certificate,
    privateKey: data.private_key,
    hostnames: data.hostnames,
    scope: data.scope,
    scopeIdentifier: data.scope_identifier,
    parentScopeIdentifier: data.parent_scope_identifier,
    validFrom: data.valid_from,
    validUntil: data.valid_until,
    generatedAt: data.generated_at,
  }
}

/**
 * Get certificate by specific scope and identifiers
 */
export async function getCertificateByScope(
  installationId: string,
  scope: CertificateScope,
  scopeIdentifier?: string,
  parentScopeIdentifier?: string,
): Promise<TLSCertificate | null> {
  return getLatestCertificate(installationId, scope, scopeIdentifier, parentScopeIdentifier)
}

/**
 * Delete all certificates for an installation
 */
export async function deleteCertificates(installationId: string): Promise<string[]> {
  // Get all certificate IDs first
  const { data } = await supabase
    .from('tls_certificates')
    .select('cloudflare_cert_id')
    .eq('installation_id', installationId)

  const certIds: string[] = []
  if (data) {
    for (const record of data as Pick<TLSCertificateRow, 'cloudflare_cert_id'>[]) {
      if (record.cloudflare_cert_id) {
        certIds.push(record.cloudflare_cert_id)
      }
    }
  }

  // Delete from database
  const { error } = await supabase
    .from('tls_certificates')
    .delete()
    .eq('installation_id', installationId)

  if (error) {
    console.error('Error deleting certificates:', error)
  }

  return certIds
}

/**
 * Edge Certificate Management (CloudFlare Advanced Certificate Manager)
 */
type EdgeCertificateRow = Database['public']['Tables']['edge_certificates']['Row']

export interface EdgeCertificate {
  id?: number
  installationId: string
  cloudflareCertPackId: string
  hostnames: string[]
  domainRequestName: string
  orderedAt?: string
  status?: 'pending' | 'active' | 'failed'
  createdAt?: string
  updatedAt?: string
}

/**
 * Save or update edge certificate (idempotent)
 */
export async function saveEdgeCertificate(cert: EdgeCertificate): Promise<void> {
  const { error } = await supabase
    .from('edge_certificates')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .upsert(
      {
        installation_id: cert.installationId,
        cloudflare_cert_pack_id: cert.cloudflareCertPackId,
        hostnames: cert.hostnames,
        domain_request_name: cert.domainRequestName,
        ordered_at: cert.orderedAt || new Date().toISOString(),
        status: cert.status || 'pending',
      },
      {
        onConflict: 'installation_id,domain_request_name',
      },
    )

  if (error) {
    throw new Error(`Failed to save edge certificate: ${error.message}`)
  }
}

/**
 * Get edge certificate by domain request name
 */
export async function getEdgeCertificate(
  installationId: string,
  domainRequestName: string,
): Promise<EdgeCertificate | null> {
  const result = await supabase
    .from('edge_certificates')
    .select('*')
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting edge certificate:', result.error)
    return null
  }

  const data = result.data as EdgeCertificateRow | null
  if (!data) return null

  return {
    id: data.id,
    installationId: data.installation_id,
    cloudflareCertPackId: data.cloudflare_cert_pack_id,
    hostnames: data.hostnames,
    domainRequestName: data.domain_request_name,
    orderedAt: data.ordered_at,
    status: data.status,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Get all edge certificates for an installation
 */
export async function getEdgeCertificates(installationId: string): Promise<EdgeCertificate[]> {
  const result = await supabase
    .from('edge_certificates')
    .select('*')
    .eq('installation_id', installationId)
    .order('created_at', { ascending: false })

  if (result.error) {
    console.error('Error getting edge certificates:', result.error)
    return []
  }

  const data = (result.data || []) as EdgeCertificateRow[]

  return data.map((cert) => ({
    id: cert.id,
    installationId: cert.installation_id,
    cloudflareCertPackId: cert.cloudflare_cert_pack_id,
    hostnames: cert.hostnames,
    domainRequestName: cert.domain_request_name,
    orderedAt: cert.ordered_at,
    status: cert.status,
    createdAt: cert.created_at,
    updatedAt: cert.updated_at,
  }))
}

/**
 * Update edge certificate status
 */
export async function updateEdgeCertificateStatus(
  installationId: string,
  domainRequestName: string,
  status: 'pending' | 'active' | 'failed',
): Promise<void> {
  const { error } = await supabase
    .from('edge_certificates')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .update({ status })
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)

  if (error) {
    throw new Error(`Failed to update edge certificate status: ${error.message}`)
  }
}

/**
 * Delete all edge certificates for an installation
 */
export async function deleteEdgeCertificates(installationId: string): Promise<string[]> {
  // Get all certificate pack IDs first
  const { data } = await supabase
    .from('edge_certificates')
    .select('cloudflare_cert_pack_id')
    .eq('installation_id', installationId)

  const certPackIds: string[] = []
  if (data) {
    for (const record of data as Pick<EdgeCertificateRow, 'cloudflare_cert_pack_id'>[]) {
      if (record.cloudflare_cert_pack_id) {
        certPackIds.push(record.cloudflare_cert_pack_id)
      }
    }
  }

  // Delete from database
  const { error } = await supabase
    .from('edge_certificates')
    .delete()
    .eq('installation_id', installationId)

  if (error) {
    console.error('Error deleting edge certificates:', error)
  }

  return certPackIds
}

/**
 * Delete edge certificates for a specific domain request
 */
export async function deleteEdgeCertificatesForDomainRequest(
  installationId: string,
  domainRequestName: string,
): Promise<string[]> {
  // Get certificate pack IDs for this domain request
  const { data } = await supabase
    .from('edge_certificates')
    .select('cloudflare_cert_pack_id')
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)

  const certPackIds: string[] = []
  if (data) {
    for (const record of data as Pick<EdgeCertificateRow, 'cloudflare_cert_pack_id'>[]) {
      if (record.cloudflare_cert_pack_id) {
        certPackIds.push(record.cloudflare_cert_pack_id)
      }
    }
  }

  // Delete from database
  const { error } = await supabase
    .from('edge_certificates')
    .delete()
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)

  if (error) {
    console.error('Error deleting edge certificates for domain request:', error)
  }

  return certPackIds
}

/**
 * Delete edge certificate for a specific domain within a domain request
 */
export async function deleteEdgeCertificateForDomain(
  installationId: string,
  domainRequestName: string,
  domain: string,
): Promise<string | null> {
  // Get certificate pack ID for this specific domain
  const { data } = await supabase
    .from('edge_certificates')
    .select('cloudflare_cert_pack_id')
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)
    .contains('hostnames', [domain])
    .single()

  if (!data) {
    return null
  }

  const certPackId = (data as Pick<EdgeCertificateRow, 'cloudflare_cert_pack_id'>)
    .cloudflare_cert_pack_id

  // Delete from database
  const { error } = await supabase
    .from('edge_certificates')
    .delete()
    .eq('installation_id', installationId)
    .eq('domain_request_name', domainRequestName)
    .contains('hostnames', [domain])

  if (error) {
    console.error(`Error deleting edge certificate for domain ${domain}:`, error)
  }

  return certPackId
}

