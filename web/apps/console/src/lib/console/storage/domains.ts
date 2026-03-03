/**
 * Domain Reservation Management
 */

import { supabase } from '../supabase'
import type { DomainReservation, DomainReservationRow } from './types'

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

  // If no reservation exists, also check the installations table directly
  // (handles cases where domain_reservation was cleaned up but installation row remains)
  if (!reservationResult.data) {
    const installationCheck = await supabase
      .from('installations')
      .select('id')
      .eq('subdomain', subdomainLower)
      .single()

    return !installationCheck.data
  }

  // Reservation exists - check if it's expired
  // @ts-expect-error — Supabase generic inference resolves to never
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
  // @ts-expect-error — Supabase generic inference resolves to never
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
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert({
      subdomain: subdomainLower,
      installation_id: installationId,
      user_id: userId,
      user_email: userEmail.toLowerCase(),
      user_name: userName,
      reserved_at: reservedAt,
      status: 'reserved' as const,
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
    // @ts-expect-error — Supabase generic inference resolves mutations to never
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

  // @ts-expect-error — Supabase generic inference resolves to never
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
  // @ts-expect-error — Supabase generic inference resolves to never
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
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert({
      subdomain: subdomainLower,
      installation_id: installationId,
      user_id: userId,
      user_email: userEmail.toLowerCase(),
      user_name: userName,
      reserved_at: reservedAt,
      status: 'reserved' as const,
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
    // @ts-expect-error — Supabase generic inference resolves mutations to never
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
