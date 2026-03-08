/**
 * Shared types for Supabase storage service
 */

import type { Database } from '../supabase-types'
import type { PiiDatabase } from '../supabase-pii-types'

export type UserRow = PiiDatabase['public']['Tables']['users']['Row']
export type OrganizationRow = Database['public']['Tables']['organizations']['Row']
export type InstallationRow = Database['public']['Tables']['installations']['Row']
export type DnsConfigurationRow = Database['public']['Tables']['dns_configurations']['Row']
export type DomainReservationRow = Database['public']['Tables']['domain_reservations']['Row']

// Org roles
export type OrgRole = 'owner' | 'admin'

// Invitation status
export type InvitationStatus = 'pending' | 'accepted' | 'rejected' | 'expired'

export interface Organization {
  id: string
  name: string
  slug: string
  createdBy: string
  createdAt: string
  updatedAt: string
}

export interface OrgMember {
  id: string
  orgId: string
  userId: string
  role: OrgRole
  addedBy: string | null
  createdAt: string
  updatedAt: string
  // Populated from PII DB join
  userEmail?: string
  userName?: string
}

export interface OrgInvitation {
  id: string
  orgId: string
  email: string
  role: 'admin'
  invitedBy: string
  status: InvitationStatus
  expiresAt: string
  createdAt: string
  updatedAt: string
  // Populated from PII DB join
  inviterName?: string
  orgName?: string
}

export interface DnsConfiguration {
  serviceName: string
  ip: string
  sshRecordId?: string | null
  routeRecordIds?: string[] // Kept for backward compatibility
  routeRecordMap?: Record<string, string> // domain -> CNAME record ID
  domainRoutes?: Array<{ domain: string }> // List of domains
}

export interface Installation {
  id: string
  orgId: string
  name?: string
  description?: string
  installationKey: string
  secretKey?: string
  setupCompleted: boolean
  subdomain?: string
  reservedAt?: string
  dnsConfigurations?: DnsConfiguration[]
  deploymentReady?: boolean
  lastHealthCheck?: string
  cloudProvider?: 'aws' | 'gcp' | 'azure' | 'oci'
  cloudLocation?: string
  deployJobExecutionName?: string
  deployJobStatus?: 'pending' | 'running' | 'succeeded' | 'failed' | 'unknown'
  deployJobStartedAt?: string
  deployJobCompletedAt?: string
  deployJobError?: string
  deployJobOperation?: 'install' | 'uninstall'
  deployJobCurrentStep?: number
  deployJobTotalSteps?: number
  deployJobStepDescription?: string
  createdAt: string
  updatedAt: string
}

export interface User {
  userId: string
  email: string
  name: string
  providers: ('github' | 'google' | 'azure-ad' | 'email')[]
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
