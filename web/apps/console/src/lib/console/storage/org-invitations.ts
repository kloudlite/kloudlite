/**
 * Organization Invitation Management
 * Uses PII database for user name lookups (cross-DB, no FK joins)
 */
/* eslint-disable @typescript-eslint/no-explicit-any */

import { supabase } from '../supabase'
import { piiSupabase } from '../supabase-pii'
import type { OrgInvitation, OrgMember } from './types'
import { addOrgMember } from './org-members'

/**
 * Create invitation for email to join an organization
 */
export async function createOrgInvitation(
  orgId: string,
  email: string,
  invitedBy: string,
): Promise<OrgInvitation> {
  const emailLower = email.toLowerCase()

  const { data, error } = await (supabase as any)
    .from('organization_invitations')
    .insert({
      org_id: orgId,
      email: emailLower,
      role: 'admin' as const,
      invited_by: invitedBy,
      status: 'pending' as const,
      expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
    })
    .select()
    .single()

  if (error || !data) {
    if (error?.code === '23505') {
      throw new Error('User already has a pending invitation')
    }
    throw new Error(`Failed to create invitation: ${error?.message ?? 'No data returned'}`)
  }

  return {
    id: data.id,
    orgId: data.org_id,
    email: data.email,
    role: data.role,
    invitedBy: data.invited_by,
    status: data.status,
    expiresAt: data.expires_at,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Get pending invitations for an organization
 */
export async function getOrgInvitations(orgId: string): Promise<OrgInvitation[]> {
  const { data, error } = await (supabase as any)
    .from('organization_invitations')
    .select(`
      *,
      organizations(name)
    `)
    .eq('org_id', orgId)
    .eq('status', 'pending')
    .order('created_at', { ascending: false })

  if (error) {
    console.error('Error getting invitations:', error)
    return []
  }

  if (!data || data.length === 0) return []

  // Get inviter names from PII DB
  const inviterIds = [...new Set(data.map((row: any) => row.invited_by))]
  const { data: inviters } = await piiSupabase
    .from('users')
    .select('user_id, name')
    .in('user_id', inviterIds)

  const inviterMap = new Map(
    (inviters || []).map((u: any) => [u.user_id, u.name]),
  )

  return data.map((row: any) => ({
    id: row.id,
    orgId: row.org_id,
    email: row.email,
    role: row.role,
    invitedBy: row.invited_by,
    status: row.status,
    expiresAt: row.expires_at,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    inviterName: inviterMap.get(row.invited_by),
    orgName: row.organizations?.name,
  }))
}

/**
 * Get pending invitations for a user's email
 */
export async function getUserPendingOrgInvitations(
  email: string,
): Promise<OrgInvitation[]> {
  const emailLower = email.toLowerCase()
  const now = new Date().toISOString()

  const { data, error } = await (supabase as any)
    .from('organization_invitations')
    .select(`
      *,
      organizations(name)
    `)
    .eq('email', emailLower)
    .eq('status', 'pending')
    .gt('expires_at', now)
    .order('created_at', { ascending: false })

  if (error) {
    console.error('Error getting user org invitations:', error)
    return []
  }

  if (!data || data.length === 0) return []

  // Get inviter names from PII DB
  const inviterIds = [...new Set(data.map((row: any) => row.invited_by))]
  const { data: inviters } = await piiSupabase
    .from('users')
    .select('user_id, name')
    .in('user_id', inviterIds)

  const inviterMap = new Map(
    (inviters || []).map((u: any) => [u.user_id, u.name]),
  )

  return data.map((row: any) => ({
    id: row.id,
    orgId: row.org_id,
    email: row.email,
    role: row.role,
    invitedBy: row.invited_by,
    status: row.status,
    expiresAt: row.expires_at,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
    inviterName: inviterMap.get(row.invited_by),
    orgName: row.organizations?.name,
  }))
}

/**
 * Accept an invitation — adds user as admin to the organization
 */
export async function acceptOrgInvitation(
  invitationId: string,
  userId: string,
  userEmail: string,
): Promise<OrgMember> {
  const { data: invitation, error: invError } = await (supabase as any)
    .from('organization_invitations')
    .select('*')
    .eq('id', invitationId)
    .single()

  if (invError || !invitation) {
    throw new Error('Invitation not found')
  }

  if (invitation.email.toLowerCase() !== userEmail.toLowerCase()) {
    throw new Error('This invitation was sent to a different email address')
  }

  if (invitation.status !== 'pending') {
    throw new Error('Invitation has already been ' + invitation.status)
  }

  if (new Date(invitation.expires_at) < new Date()) {
    throw new Error('Invitation has expired')
  }

  // Add member to org
  const member = await addOrgMember(
    invitation.org_id,
    userId,
    'admin',
    invitation.invited_by,
  )

  // Mark invitation as accepted
  await (supabase as any)
    .from('organization_invitations')
    .update({ status: 'accepted' })
    .eq('id', invitationId)

  return member
}

/**
 * Reject an invitation
 */
export async function rejectOrgInvitation(
  invitationId: string,
  userEmail: string,
): Promise<void> {
  const { data: invitation, error: invError } = await (supabase as any)
    .from('organization_invitations')
    .select('*')
    .eq('id', invitationId)
    .single()

  if (invError || !invitation) {
    throw new Error('Invitation not found')
  }

  if (invitation.email.toLowerCase() !== userEmail.toLowerCase()) {
    throw new Error('This invitation was sent to a different email address')
  }

  const { error } = await (supabase as any)
    .from('organization_invitations')
    .update({ status: 'rejected' })
    .eq('id', invitationId)

  if (error) {
    throw new Error(`Failed to reject invitation: ${error.message}`)
  }
}

/**
 * Delete/cancel an invitation
 */
export async function deleteOrgInvitation(invitationId: string): Promise<void> {
  const { error } = await (supabase as any)
    .from('organization_invitations')
    .delete()
    .eq('id', invitationId)

  if (error) {
    throw new Error(`Failed to delete invitation: ${error.message}`)
  }
}
