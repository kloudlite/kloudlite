/**
 * Shared types for Supabase storage service
 */

import type { Database } from '../supabase-types'

export type UserRegistrationRow = Database['public']['Tables']['user_registrations']['Row']
export type InstallationRow = Database['public']['Tables']['installations']['Row']
export type IPRecordRow = Database['public']['Tables']['ip_records']['Row']
export type DomainReservationRow = Database['public']['Tables']['domain_reservations']['Row']

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
  cloudProvider?: 'aws' | 'gcp' | 'azure' | 'oci'
  cloudLocation?: string
  acaJobExecutionName?: string
  acaJobStatus?: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown'
  acaJobStartedAt?: string
  acaJobCompletedAt?: string
  acaJobError?: string
  acaJobOperation?: 'install' | 'uninstall'
  acaJobCurrentStep?: number
  acaJobTotalSteps?: number
  acaJobStepDescription?: string
  createdAt: string
  updatedAt: string
}

export interface UserRegistration {
  userId: string
  email: string
  name: string
  providers: ('github' | 'google' | 'azure-ad' | 'email')[]
  registeredAt: string
  createdAt: string
  updatedAt: string
}

export interface MagicLinkToken {
  id: string
  email: string
  token: string
  expiresAt: string
  usedAt: string | null
  createdAt: string
  ipAddress: string | null
  userAgent: string | null
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
