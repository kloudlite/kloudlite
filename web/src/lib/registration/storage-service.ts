/**
 * Storage Service using Cloudflare KV
 *
 * Handles all data persistence for user registrations and domain reservations
 */

import { kvGet, kvPut, kvList, kvDelete } from './cloudflare-kv'

export interface IPRecord {
  type: 'installation' | 'workmachine'
  ip: string // A record IP address
  workMachineName?: string // For workmachine only - extracted from domain pattern (e.g., "user1" from "user1.test.khost.dev")
  configuredAt: string // When this IP was configured
}

export interface UserRegistration {
  userId: string // OAuth provider ID
  email: string
  name: string
  provider: 'github' | 'google' | 'azure-ad'
  registeredAt: string
  installationKey: string // Unique key for installation verification
  secretKey?: string // Secret key for bearer token authentication (generated on first deployment verification)
  hasCompletedInstallation: boolean
  subdomain?: string
  reservedAt?: string
  ipRecords?: IPRecord[] // All IP records with domain and A record
  deploymentReady?: boolean // Flag set when IPs are configured
  lastHealthCheck?: string // Last time deployment checked in (polling)
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
  const key = `user:${email.toLowerCase()}`
  return await kvGet<UserRegistration>(key)
}

/**
 * Get user registration by user ID (for backwards compatibility)
 */
export async function getUserRegistration(userId: string): Promise<UserRegistration | null> {
  // This is now just a wrapper - we don't use userId as primary key anymore
  // In practice, use getUserByEmail instead
  return null
}

/**
 * Create or update user registration
 * Email is the primary key - one email = one registration
 */
export async function saveUserRegistration(registration: UserRegistration): Promise<void> {
  // Use email as primary key
  const key = `user:${registration.email.toLowerCase()}`
  await kvPut(key, registration)

  // Create an index by installation key for lookups
  if (registration.installationKey) {
    const installKeyIndex = `installkey:${registration.installationKey}`
    await kvPut(installKeyIndex, registration.email.toLowerCase())
  }
}

/**
 * Mark user as having completed installation
 */
export async function markInstallationComplete(email: string): Promise<void> {
  const registration = await getUserByEmail(email)
  if (!registration) {
    throw new Error('User registration not found')
  }

  registration.hasCompletedInstallation = true
  await saveUserRegistration(registration)
}

/**
 * Check if subdomain is available
 */
export async function isSubdomainAvailable(subdomain: string): Promise<boolean> {
  const key = `domain:${subdomain.toLowerCase()}`
  const reservation = await kvGet<DomainReservation>(key)

  // Check if subdomain is reserved
  if (reservation) {
    return false
  }

  // Check against reserved keywords
  const RESERVED_KEYWORDS = [
    'www', 'api', 'admin', 'mail', 'smtp', 'ftp', 'ssh', 'vpn',
    'dev', 'staging', 'prod', 'production', 'test', 'demo',
    'app', 'portal', 'dashboard', 'console', 'docs', 'blog',
    'status', 'support', 'help', 'cdn', 'static', 'assets',
  ]

  if (RESERVED_KEYWORDS.includes(subdomain.toLowerCase())) {
    return false
  }

  return true
}

/**
 * Reserve a subdomain for a user with atomic lock
 */
export async function reserveSubdomain(
  subdomain: string,
  userId: string,
  userEmail: string,
  userName: string
): Promise<DomainReservation> {
  const subdomainLower = subdomain.toLowerCase()
  const domainKey = `domain:${subdomainLower}`
  const lockKey = `lock:domain:${subdomainLower}`

  // Try to acquire lock (30 second TTL)
  try {
    // First check if domain already exists
    const existing = await getDomainReservation(subdomainLower)
    if (existing) {
      throw new Error('Subdomain is already reserved')
    }

    // Create lock entry with 30s TTL to prevent concurrent reservations
    await kvPut(lockKey, { lockedBy: userId, lockedAt: new Date().toISOString() }, 30)

    // Double-check availability after acquiring lock
    const stillAvailable = await isSubdomainAvailable(subdomainLower)
    if (!stillAvailable) {
      await kvDelete(lockKey) // Release lock
      throw new Error('Subdomain was reserved by another user')
    }

    const reservation: DomainReservation = {
      subdomain: subdomainLower,
      userId,
      reservedAt: new Date().toISOString(),
      status: 'reserved',
      userEmail,
      userName,
    }

    // Save domain reservation
    await kvPut(domainKey, reservation)

    // Update user registration with subdomain (use email as key)
    const userRegistration = await getUserByEmail(userEmail)
    if (userRegistration) {
      userRegistration.subdomain = subdomainLower
      userRegistration.reservedAt = reservation.reservedAt
      await saveUserRegistration(userRegistration)
    }

    // Release lock
    await kvDelete(lockKey)

    return reservation
  } catch (error) {
    // Always try to release lock on error
    try {
      await kvDelete(lockKey)
    } catch (lockError) {
      // Ignore lock deletion errors (lock will expire anyway)
    }
    throw error
  }
}

/**
 * Get domain reservation by subdomain
 */
export async function getDomainReservation(subdomain: string): Promise<DomainReservation | null> {
  const key = `domain:${subdomain.toLowerCase()}`
  return await kvGet<DomainReservation>(key)
}

/**
 * List all reserved domains
 */
export async function listReservedDomains(): Promise<string[]> {
  const keys = await kvList('domain:')
  return keys.map(key => key.replace('domain:', ''))
}

/**
 * Get user by installation key
 */
export async function getUserByInstallationKey(installationKey: string): Promise<UserRegistration | null> {
  const installKeyIndex = `installkey:${installationKey}`
  const email = await kvGet<string>(installKeyIndex)

  if (!email) {
    return null
  }

  const user = await getUserByEmail(email)

  // Secret key will be generated on first deployment verification (in verify-key API)
  // Do not auto-generate here

  return user
}
