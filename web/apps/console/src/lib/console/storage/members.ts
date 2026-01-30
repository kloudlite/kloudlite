/**
 * Team Member Management
 */

import { supabase } from '../supabase'
import type { MemberRole, InstallationMember } from './types'

/**
 * Get member's role for an installation
 * Returns null if not a member
 */
export async function getMemberRole(
  installationId: string,
  userId: string
): Promise<MemberRole | null> {
  const { data, error } = await (supabase as any)
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

  if (!data) return null

  return (data as { role: string }).role as MemberRole
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
  const { data, error } = await (supabase as any)
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

  if (!data) {
    throw new Error('Failed to add member: No data returned')
  }

  const memberData = data as any

  return {
    id: memberData.id,
    installationId: memberData.installation_id,
    userId: memberData.user_id,
    role: memberData.role,
    addedBy: memberData.added_by,
    addedAt: memberData.added_at,
    createdAt: memberData.created_at,
    updatedAt: memberData.updated_at,
  }
}

/**
 * Update member role
 */
export async function updateMemberRole(
  memberId: string,
  newRole: MemberRole
): Promise<void> {
  const { error } = await (supabase as any)
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
