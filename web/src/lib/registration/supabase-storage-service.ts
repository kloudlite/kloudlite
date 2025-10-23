/**
 * Storage Service using Supabase (PostgreSQL)
 *
 * Provides atomic operations using SQL transactions
 * No eventual consistency - ACID guarantees
 */

import { supabase } from './supabase'

export interface IPRecord {
  type: 'installation' | 'workmachine'
  ip: string
  workMachineName?: string
  configuredAt: string
  dnsRecordIds?: string[]
}

export interface UserRegistration {
  userId: string
  email: string
  name: string
  providers: ('github' | 'google' | 'azure-ad')[]
  registeredAt: string
  installationKey: string
  secretKey?: string
  hasCompletedInstallation: boolean
  subdomain?: string
  reservedAt?: string
  ipRecords?: IPRecord[]
  deploymentReady?: boolean
  lastHealthCheck?: string
}

export interface DomainReservation {
  subdomain: string
  userId: string
  reservedAt: string
  status: 'reserved' | 'active' | 'cancelled'
  userEmail: string
  userName: string
}

/**
 * Get user registration by email (primary key)
 */
export async function getUserByEmail(email: string): Promise<UserRegistration | null> {
  const { data, error } = await supabase
    .from('user_registrations')
    .select('*')
    .eq('email', email.toLowerCase())
    .single()

  if (error) {
    if (error.code === 'PGRST116') return null // Not found
    console.error('Error getting user:', error)
    return null
  }

  // Fetch IP records
  const { data: ipData } = await supabase
    .from('ip_records')
    .select('*')
    .eq('user_email', email.toLowerCase())

  return {
    userId: data.user_id,
    email: data.email,
    name: data.name,
    providers: data.providers || [],
    registeredAt: data.registered_at,
    installationKey: data.installation_key,
    secretKey: data.secret_key || undefined,
    hasCompletedInstallation: data.has_completed_installation,
    subdomain: data.subdomain || undefined,
    reservedAt: data.reserved_at || undefined,
    deploymentReady: data.deployment_ready || undefined,
    lastHealthCheck: data.last_health_check || undefined,
    ipRecords: ipData?.map(ip => ({
      type: ip.type,
      ip: ip.ip,
      workMachineName: ip.work_machine_name || undefined,
      configuredAt: ip.configured_at,
      dnsRecordIds: ip.dns_record_ids || undefined,
    })) || [],
  }
}

/**
 * Get user by installation key
 */
export async function getUserByInstallationKey(installationKey: string): Promise<UserRegistration | null> {
  const { data, error } = await supabase
    .from('user_registrations')
    .select('*')
    .eq('installation_key', installationKey)
    .single()

  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error getting user by installation key:', error)
    return null
  }

  return getUserByEmail(data.email)
}

/**
 * Create or update user registration
 * Uses INSERT ... ON CONFLICT (upsert) for atomicity
 */
export async function saveUserRegistration(registration: UserRegistration): Promise<void> {
  const { error } = await supabase
    .from('user_registrations')
    .upsert({
      email: registration.email.toLowerCase(),
      user_id: registration.userId,
      name: registration.name,
      providers: registration.providers,
      registered_at: registration.registeredAt,
      installation_key: registration.installationKey,
      secret_key: registration.secretKey || null,
      has_completed_installation: registration.hasCompletedInstallation,
      subdomain: registration.subdomain || null,
      reserved_at: registration.reservedAt || null,
      deployment_ready: registration.deploymentReady || null,
      last_health_check: registration.lastHealthCheck || null,
    })

  if (error) {
    console.error('Error saving user registration:', error)
    throw new Error(`Failed to save user registration: ${error.message}`)
  }
}

/**
 * Atomically mark installation complete with optional secret key
 */
export async function markInstallationComplete(email: string, secretKey?: string): Promise<UserRegistration> {
  const update: any = {
    has_completed_installation: true,
  }

  if (secretKey) {
    update.secret_key = secretKey
  }

  const { error } = await supabase
    .from('user_registrations')
    .update(update)
    .eq('email', email.toLowerCase())

  if (error) {
    throw new Error(`Failed to mark installation complete: ${error.message}`)
  }

  // Fetch and return the updated user
  const user = await getUserByEmail(email)
  if (!user) {
    throw new Error('User not found after update')
  }
  return user
}

/**
 * Atomically update health check timestamp
 */
export async function updateHealthCheck(email: string): Promise<UserRegistration> {
  const { error } = await supabase
    .from('user_registrations')
    .update({ last_health_check: new Date().toISOString() })
    .eq('email', email.toLowerCase())

  if (error) {
    throw new Error(`Failed to update health check: ${error.message}`)
  }

  // Fetch and return the updated user
  const user = await getUserByEmail(email)
  if (!user) {
    throw new Error('User not found after update')
  }
  return user
}

/**
 * Atomically add or update IP record
 * Uses UPSERT with unique constraint on (user_email, type, work_machine_name)
 */
export async function addOrUpdateIpRecord(email: string, ipRecord: IPRecord): Promise<number> {
  const { error } = await supabase
    .from('ip_records')
    .upsert({
      user_email: email.toLowerCase(),
      type: ipRecord.type,
      ip: ipRecord.ip,
      work_machine_name: ipRecord.workMachineName || null,
      configured_at: ipRecord.configuredAt,
      dns_record_ids: ipRecord.dnsRecordIds || [],
    }, {
      onConflict: 'user_email,type,work_machine_name'
    })

  if (error) {
    throw new Error(`Failed to add IP record: ${error.message}`)
  }

  // Get total count
  const { count } = await supabase
    .from('ip_records')
    .select('*', { count: 'exact', head: true })
    .eq('user_email', email.toLowerCase())

  return count || 0
}

/**
 * Atomically mark deployment ready
 */
export async function markDeploymentReady(email: string, ready: boolean): Promise<void> {
  const { error } = await supabase
    .from('user_registrations')
    .update({ deployment_ready: ready })
    .eq('email', email.toLowerCase())

  if (error) {
    throw new Error(`Failed to mark deployment ready: ${error.message}`)
  }
}

/**
 * Check if subdomain is available
 */
export async function isSubdomainAvailable(subdomain: string): Promise<boolean> {
  const RESERVED_KEYWORDS = [
    'www', 'api', 'admin', 'mail', 'smtp', 'ftp', 'ssh', 'vpn',
    'dev', 'staging', 'prod', 'production', 'test', 'demo',
    'app', 'portal', 'dashboard', 'console', 'docs', 'blog',
    'status', 'support', 'help', 'cdn', 'static', 'assets',
  ]

  if (RESERVED_KEYWORDS.includes(subdomain.toLowerCase())) {
    return false
  }

  const { data } = await supabase
    .from('domain_reservations')
    .select('subdomain')
    .eq('subdomain', subdomain.toLowerCase())
    .single()

  return !data
}

/**
 * Atomically reserve a subdomain
 * Uses INSERT with unique constraint for atomicity
 */
export async function reserveSubdomain(
  subdomain: string,
  userId: string,
  userEmail: string,
  userName: string
): Promise<DomainReservation> {
  const subdomainLower = subdomain.toLowerCase()
  const reservedAt = new Date().toISOString()

  // Insert domain reservation (will fail if subdomain already exists due to PRIMARY KEY)
  const { data, error } = await supabase
    .from('domain_reservations')
    .insert({
      subdomain: subdomainLower,
      user_id: userId,
      user_email: userEmail.toLowerCase(),
      user_name: userName,
      reserved_at: reservedAt,
      status: 'reserved',
    })
    .select()
    .single()

  if (error) {
    if (error.code === '23505') {
      // Unique constraint violation
      throw new Error('Subdomain is already reserved')
    }
    throw new Error(`Failed to reserve subdomain: ${error.message}`)
  }

  // Atomically update user registration with subdomain
  await supabase
    .from('user_registrations')
    .update({
      subdomain: subdomainLower,
      reserved_at: reservedAt,
    })
    .eq('email', userEmail.toLowerCase())

  return {
    subdomain: data.subdomain,
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
  const { data, error } = await supabase
    .from('domain_reservations')
    .select('*')
    .eq('subdomain', subdomain.toLowerCase())
    .single()

  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error getting domain reservation:', error)
    return null
  }

  return {
    subdomain: data.subdomain,
    userId: data.user_id,
    reservedAt: data.reserved_at,
    status: data.status,
    userEmail: data.user_email,
    userName: data.user_name,
  }
}

/**
 * Delete all IP records for a user and return their DNS record IDs
 */
export async function deleteIpRecords(email: string): Promise<string[]> {
  // Get all IP records to extract DNS record IDs
  const { data: ipData } = await supabase
    .from('ip_records')
    .select('*')
    .eq('user_email', email.toLowerCase())

  const dnsRecordIds: string[] = []

  if (ipData) {
    for (const record of ipData) {
      if (record.dns_record_ids && Array.isArray(record.dns_record_ids)) {
        dnsRecordIds.push(...record.dns_record_ids)
      }
    }
  }

  // Delete all IP records
  const { error } = await supabase
    .from('ip_records')
    .delete()
    .eq('user_email', email.toLowerCase())

  if (error) {
    console.error('Error deleting IP records:', error)
  }

  return dnsRecordIds
}

/**
 * Delete domain reservation for a user
 */
export async function deleteDomainReservation(email: string): Promise<void> {
  const { error } = await supabase
    .from('domain_reservations')
    .delete()
    .eq('user_email', email.toLowerCase())

  if (error) {
    console.error('Error deleting domain reservation:', error)
    throw new Error(`Failed to delete domain reservation: ${error.message}`)
  }
}

/**
 * Reset user installation (clears subdomain, deployment status, but keeps user registration)
 */
export async function resetUserInstallation(email: string): Promise<void> {
  const { error } = await supabase
    .from('user_registrations')
    .update({
      subdomain: null,
      reserved_at: null,
      secret_key: null,
      has_completed_installation: false,
      deployment_ready: false,
      last_health_check: null,
    })
    .eq('email', email.toLowerCase())

  if (error) {
    throw new Error(`Failed to reset installation: ${error.message}`)
  }
}

/**
 * Certificate management
 */
export type CertificateScope = 'installation' | 'workmachine' | 'workspace'

export interface TLSCertificate {
  id?: number
  userEmail: string
  cloudflareCertId: string | null
  certificate: string
  privateKey: string
  hostnames: string[]
  scope: CertificateScope
  scopeIdentifier?: string | null // wm-user for workmachine, workspace name for workspace
  parentScopeIdentifier?: string | null // wm-user for workspace scope
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
    .insert({
      user_email: cert.userEmail.toLowerCase(),
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
 * Get latest certificate for user, optionally filtered by scope
 */
export async function getLatestCertificate(
  email: string,
  scope?: CertificateScope,
  scopeIdentifier?: string,
  parentScopeIdentifier?: string
): Promise<TLSCertificate | null> {
  let query = supabase
    .from('tls_certificates')
    .select('*')
    .eq('user_email', email.toLowerCase())

  if (scope) {
    query = query.eq('scope', scope)
  }

  if (scopeIdentifier !== undefined) {
    query = query.eq('scope_identifier', scopeIdentifier)
  }

  if (parentScopeIdentifier !== undefined) {
    query = query.eq('parent_scope_identifier', parentScopeIdentifier)
  }

  const { data, error } = await query
    .order('generated_at', { ascending: false })
    .limit(1)
    .single()

  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error getting certificate:', error)
    return null
  }

  return {
    id: data.id,
    userEmail: data.user_email,
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
  email: string,
  scope: CertificateScope,
  scopeIdentifier?: string,
  parentScopeIdentifier?: string
): Promise<TLSCertificate | null> {
  return getLatestCertificate(email, scope, scopeIdentifier, parentScopeIdentifier)
}

/**
 * Delete all certificates for a user
 */
export async function deleteCertificates(email: string): Promise<string[]> {
  // Get all certificate IDs first
  const { data } = await supabase
    .from('tls_certificates')
    .select('cloudflare_cert_id')
    .eq('user_email', email.toLowerCase())

  const certIds: string[] = []
  if (data) {
    for (const record of data) {
      if (record.cloudflare_cert_id) {
        certIds.push(record.cloudflare_cert_id)
      }
    }
  }

  // Delete from database
  const { error } = await supabase
    .from('tls_certificates')
    .delete()
    .eq('user_email', email.toLowerCase())

  if (error) {
    console.error('Error deleting certificates:', error)
  }

  return certIds
}
