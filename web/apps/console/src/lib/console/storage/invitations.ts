/**
 * Invitation Management
 */

import { supabase } from '../supabase'
import type { MemberRole, InstallationInvitation, InstallationMember } from './types'
import { addInstallationMember } from './members'

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

  const { data, error} = await (supabase as any)
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
  const { data, error } = await (supabase as any)
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

  const { data, error } = await (supabase as any)
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
  const { data: invitation, error: invError } = await (supabase as any)
    .from('installation_invitations')
    .select('*')
    .eq('id', invitationId)
    .single()

  if (invError || !invitation) {
    throw new Error('Invitation not found')
  }

  const invData = invitation as any

  // Check if expired
  if (new Date(invData.expires_at) < new Date()) {
    throw new Error('Invitation has expired')
  }

  // Add member
  const member = await addInstallationMember(
    invData.installation_id,
    userId,
    invData.role,
    invData.invited_by
  )

  // Mark invitation as accepted
  await (supabase as any)
    .from('installation_invitations')
    .update({ status: 'accepted' })
    .eq('id', invitationId)

  return member
}

/**
 * Reject an invitation
 */
export async function rejectInvitation(invitationId: string): Promise<void> {
  const { error } = await (supabase as any)
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
  const { error } = await (supabase as any)
    .from('installation_invitations')
    .delete()
    .eq('id', invitationId)

  if (error) {
    throw new Error(`Failed to delete invitation: ${error.message}`)
  }
}
