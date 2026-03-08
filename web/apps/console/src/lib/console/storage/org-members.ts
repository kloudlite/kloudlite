/**
 * Organization Member Management
 * Two roles: owner (billing + members + installations) and admin (members + installations)
 */
/* eslint-disable @typescript-eslint/no-explicit-any */

import { supabase } from '../supabase'
import { piiSupabase } from '../supabase-pii'
import type { OrgRole, OrgMember } from './types'

/**
 * Get member's role for an organization
 * Returns null if not a member
 */
export async function getOrgMemberRole(
  orgId: string,
  userId: string,
): Promise<OrgRole | null> {
  const { data, error } = await (supabase as any)
    .from('organization_members')
    .select('role')
    .eq('org_id', orgId)
    .eq('user_id', userId)
    .single()

  if (error) {
    if (error.code !== 'PGRST116') {
      console.error('Error getting org member role:', error)
    }
    return null
  }

  return data?.role as OrgRole
}

/**
 * Check if user is a member of the organization
 */
export async function isOrgMember(orgId: string, userId: string): Promise<boolean> {
  const role = await getOrgMemberRole(orgId, userId)
  return role !== null
}

/**
 * Check if user is the owner of the organization
 */
export async function isOrgOwner(orgId: string, userId: string): Promise<boolean> {
  const role = await getOrgMemberRole(orgId, userId)
  return role === 'owner'
}

/**
 * Get all members for an organization with user details from PII DB
 */
export async function getOrgMembers(orgId: string): Promise<OrgMember[]> {
  const { data: membersData, error } = await supabase
    .from('organization_members')
    .select('*')
    .eq('org_id', orgId)
    .order('created_at', { ascending: true })

  if (error) {
    console.error('Error getting organization members:', error)
    return []
  }

  const members = (membersData || []) as any[]
  if (members.length === 0) return []

  // Get unique user IDs and fetch details from PII DB
  const userIds = [...new Set(members.map((m: any) => m.user_id))]
  const { data: usersData, error: usersError } = await piiSupabase
    .from('users')
    .select('user_id, email, name')
    .in('user_id', userIds)

  if (usersError) {
    console.error('Error getting user details:', usersError)
  }

  const userMap = new Map(
    (usersData || []).map((u: any) => [u.user_id, u]),
  )

  return members.map((row: any) => {
    const user = userMap.get(row.user_id) || {}
    return {
      id: row.id,
      orgId: row.org_id,
      userId: row.user_id,
      role: row.role as OrgRole,
      addedBy: row.added_by,
      createdAt: row.created_at,
      updatedAt: row.updated_at,
      userEmail: user.email || 'Unknown',
      userName: user.name || 'Unknown User',
    }
  })
}

/**
 * Add a member to an organization
 */
export async function addOrgMember(
  orgId: string,
  userId: string,
  role: OrgRole,
  addedBy: string,
): Promise<OrgMember> {
  const { data, error } = await (supabase as any)
    .from('organization_members')
    .insert({
      org_id: orgId,
      user_id: userId,
      role,
      added_by: addedBy,
    })
    .select()
    .single()

  if (error) {
    if (error.code === '23505') {
      throw new Error('User is already a member of this organization')
    }
    throw new Error(`Failed to add member: ${error.message}`)
  }

  return {
    id: data.id,
    orgId: data.org_id,
    userId: data.user_id,
    role: data.role as OrgRole,
    addedBy: data.added_by,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Remove a member from organization
 * Cannot remove the owner
 */
export async function removeOrgMember(memberId: string): Promise<void> {
  // Check if this is the owner
  const { data: member } = await (supabase as any)
    .from('organization_members')
    .select('role')
    .eq('id', memberId)
    .single()

  if (member?.role === 'owner') {
    throw new Error('Cannot remove the organization owner')
  }

  const { error } = await supabase
    .from('organization_members')
    .delete()
    .eq('id', memberId)

  if (error) {
    throw new Error(`Failed to remove member: ${error.message}`)
  }
}

/**
 * Transfer ownership to another member
 * Demotes current owner to admin, promotes target to owner
 */
export async function transferOwnership(
  orgId: string,
  currentOwnerId: string,
  newOwnerId: string,
): Promise<void> {
  const { error } = await (supabase as any).rpc('transfer_org_ownership', {
    p_org_id: orgId,
    p_old_owner: currentOwnerId,
    p_new_owner: newOwnerId,
  })

  if (!error) return

  // Fall back to two-step update if RPC doesn't exist yet
  if (error.code === '42883' || error.message?.includes('function') || error.code === 'PGRST202') {
    const { error: demoteError } = await (supabase as any)
      .from('organization_members')
      .update({ role: 'admin', updated_at: new Date().toISOString() })
      .eq('org_id', orgId)
      .eq('user_id', currentOwnerId)
      .eq('role', 'owner')

    if (demoteError) throw new Error(`Failed to demote current owner: ${demoteError.message}`)

    const { error: promoteError } = await (supabase as any)
      .from('organization_members')
      .update({ role: 'owner', updated_at: new Date().toISOString() })
      .eq('org_id', orgId)
      .eq('user_id', newOwnerId)

    if (promoteError) throw new Error(`Failed to promote new owner: ${promoteError.message}`)
    return
  }

  throw new Error(`Failed to transfer ownership: ${error.message}`)
}
