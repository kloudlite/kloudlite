/**
 * Organization Management Functions
 */

import type { Database } from '../supabase-types'
import { supabase } from '../supabase'
import type { Organization, OrganizationRow } from './types'
import { ensureCreditAccount } from './credits'

function mapToOrganization(row: OrganizationRow): Organization {
  return {
    id: row.id,
    name: row.name,
    slug: row.slug,
    createdBy: row.created_by,
    createdAt: row.created_at,
    updatedAt: row.updated_at,
  }
}

/**
 * Validate slug format (DNS-1123 label: lowercase alphanumeric + hyphens, 3-63 chars)
 */
export function isValidSlug(slug: string): boolean {
  return /^[a-z][a-z0-9-]{1,61}[a-z0-9]$/.test(slug)
}

/**
 * Get organization by ID
 */
export async function getOrganizationById(orgId: string): Promise<Organization | null> {
  const { data, error } = await supabase
    .from('organizations')
    .select('*')
    .eq('id', orgId)
    .single()

  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error getting organization:', error)
    return null
  }

  return mapToOrganization(data as OrganizationRow)
}

/**
 * Get organization by slug
 */
export async function getOrganizationBySlug(slug: string): Promise<Organization | null> {
  const { data, error } = await supabase
    .from('organizations')
    .select('*')
    .eq('slug', slug)
    .single()

  if (error) {
    if (error.code === 'PGRST116') return null
    console.error('Error getting organization by slug:', error)
    return null
  }

  return mapToOrganization(data as OrganizationRow)
}

/**
 * Check if a slug is available
 */
export async function isSlugAvailable(slug: string): Promise<boolean> {
  const { count, error } = await supabase
    .from('organizations')
    .select('*', { count: 'exact', head: true })
    .eq('slug', slug)

  if (error) return false
  return (count ?? 0) === 0
}

/**
 * Get all organizations a user belongs to
 */
export async function getUserOrganizations(userId: string): Promise<Organization[]> {
  // Get org IDs where user is a member
  const { data: memberData, error: memberError } = await supabase
    .from('organization_members')
    .select('org_id')
    .eq('user_id', userId)

  if (memberError || !memberData || memberData.length === 0) {
    return []
  }

  const orgIds = memberData.map((m: { org_id: string }) => m.org_id)

  const { data, error } = await supabase
    .from('organizations')
    .select('*')
    .in('id', orgIds)
    .order('created_at', { ascending: false })

  if (error) {
    console.error('Error getting user organizations:', error)
    return []
  }

  return (data as OrganizationRow[]).map(mapToOrganization)
}

/**
 * Create a new organization and add the creator as owner
 */
export async function createOrganization(
  userId: string,
  name: string,
  slug: string,
): Promise<Organization> {
  if (!isValidSlug(slug)) {
    throw new Error('Invalid slug: must be lowercase alphanumeric with hyphens, 3-63 characters')
  }

  // Try atomic RPC first; fall back to two-step insert if RPC doesn't exist yet
  const { data: rpcData, error: rpcError } = await (supabase as any).rpc('create_organization_with_owner', {
    p_name: name,
    p_slug: slug,
    p_created_by: userId,
  })

  if (!rpcError) {
    const orgId = rpcData as string
    const org = await getOrganizationById(orgId)
    if (!org) throw new Error('Organization not found after creation')

    // Create a zero-balance credit account for the new org
    try {
      await ensureCreditAccount(orgId)
    } catch (err) {
      console.error('Failed to create credit account for new org:', err)
    }

    return org
  }

  // If RPC doesn't exist (not yet migrated), fall back to two-step insert
  if (rpcError.code === '42883' || rpcError.message?.includes('function') || rpcError.code === 'PGRST202') {
    type OrgInsert = Database['public']['Tables']['organizations']['Insert']
    const insertData: OrgInsert = { name, slug, created_by: userId }

    const { data, error } = await supabase
      .from('organizations')
      // @ts-expect-error — Supabase generic inference resolves mutations to never
      .insert(insertData)
      .select()
      .single()

    if (error) {
      if (error.code === '23505') {
        throw new Error('An organization with this slug already exists')
      }
      throw new Error(`Failed to create organization: ${error.message}`)
    }

    const org = mapToOrganization(data as OrganizationRow)

    // Add creator as owner
    const { error: memberError } = await supabase
      .from('organization_members')
      // @ts-expect-error — Supabase generic inference resolves mutations to never
      .insert({
        org_id: org.id,
        user_id: userId,
        role: 'owner' as const,
        added_by: userId,
      })

    if (memberError) {
      console.error('Error adding owner to organization_members:', memberError)
    }

    // Create a zero-balance credit account for the new org
    try {
      await ensureCreditAccount(org.id)
    } catch (err) {
      console.error('Failed to create credit account for new org:', err)
    }

    return org
  }

  // Actual error from the RPC
  if (rpcError.code === '23505') {
    throw new Error('An organization with this slug already exists')
  }
  throw new Error(`Failed to create organization: ${rpcError.message}`)
}

/**
 * Update organization name
 */
export async function updateOrganization(
  orgId: string,
  updates: { name?: string },
): Promise<Organization> {
  type OrgUpdate = Database['public']['Tables']['organizations']['Update']
  const updateData: OrgUpdate = {}
  if (updates.name !== undefined) updateData.name = updates.name

  const { error } = await supabase
    .from('organizations')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .update(updateData)
    .eq('id', orgId)

  if (error) {
    throw new Error(`Failed to update organization: ${error.message}`)
  }

  const org = await getOrganizationById(orgId)
  if (!org) throw new Error('Organization not found after update')
  return org
}

/**
 * Delete an organization (cascades to members, invitations, installations, billing)
 */
export async function deleteOrganization(orgId: string): Promise<void> {
  const { error } = await supabase
    .from('organizations')
    .delete()
    .eq('id', orgId)

  if (error) {
    throw new Error(`Failed to delete organization: ${error.message}`)
  }
}
