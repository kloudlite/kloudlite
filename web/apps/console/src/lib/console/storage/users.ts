/**
 * User Management Functions
 * Uses PII database for user data storage
 */

import type { PiiDatabase } from '../supabase-pii-types'
import { piiSupabase } from '../supabase-pii'
import type { User, UserRow } from './types'

/**
 * Get user by userId
 */
export async function getUserById(userId: string): Promise<User | null> {
  const result = await piiSupabase
    .from('users')
    .select('*')
    .eq('user_id', userId)
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting user:', result.error)
    return null
  }

  const data = result.data as UserRow | null
  if (!data) return null

  return {
    userId: data.user_id,
    email: data.email,
    name: data.name,
    providers: data.providers || [],
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Get user by email
 */
export async function getUserByEmail(email: string): Promise<User | null> {
  const result = await piiSupabase
    .from('users')
    .select('*')
    .eq('email', email.toLowerCase())
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting user:', result.error)
    return null
  }

  const data = result.data as UserRow | null
  if (!data) return null

  return {
    userId: data.user_id,
    email: data.email,
    name: data.name,
    providers: data.providers || [],
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Create or update user
 */
export async function saveUser(user: User): Promise<void> {
  type UserInsert = PiiDatabase['public']['Tables']['users']['Insert']
  type UserUpdate = PiiDatabase['public']['Tables']['users']['Update']

  const insertData: UserInsert = {
    user_id: user.userId,
    email: user.email.toLowerCase(),
    name: user.name,
    providers: user.providers,
  }

  // Try to insert first
  const { error: insertError } = await piiSupabase
    .from('users')
    // @ts-expect-error — Supabase generic inference resolves mutations to never
    .insert(insertData)

  // If user already exists (unique constraint violation), update instead
  if (insertError && insertError.code === '23505') {
    const updateData: UserUpdate = {
      name: user.name,
      providers: user.providers,
    }

    const { error: updateError } = await piiSupabase
      .from('users')
      // @ts-expect-error — Supabase generic inference resolves mutations to never
      .update(updateData)
      .eq('user_id', user.userId)

    if (updateError) {
      console.error('Error updating user:', updateError)
      throw new Error(`Failed to update user: ${updateError.message}`)
    }
  } else if (insertError) {
    console.error('Error saving user:', insertError)
    throw new Error(`Failed to save user: ${insertError.message}`)
  }
}
