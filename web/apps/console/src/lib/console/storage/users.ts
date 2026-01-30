/**
 * User Management Functions
 */

import type { Database } from '../supabase-types'
import { supabase } from '../supabase'
import type { UserRegistration, UserRegistrationRow } from './types'

/**
 * Get user by userId
 */
export async function getUserById(userId: string): Promise<UserRegistration | null> {
  const result = await supabase
    .from('user_registrations')
    .select('*')
    .eq('user_id', userId)
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting user:', result.error)
    return null
  }

  const data = result.data as UserRegistrationRow | null
  if (!data) return null

  return {
    userId: data.user_id,
    email: data.email,
    name: data.name,
    providers: data.providers || [],
    registeredAt: data.registered_at,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Get user by email
 */
export async function getUserByEmail(email: string): Promise<UserRegistration | null> {
  const result = await supabase
    .from('user_registrations')
    .select('*')
    .eq('email', email.toLowerCase())
    .single()

  if (result.error) {
    if (result.error.code === 'PGRST116') return null
    console.error('Error getting user:', result.error)
    return null
  }

  const data = result.data as UserRegistrationRow | null
  if (!data) return null

  return {
    userId: data.user_id,
    email: data.email,
    name: data.name,
    providers: data.providers || [],
    registeredAt: data.registered_at,
    createdAt: data.created_at,
    updatedAt: data.updated_at,
  }
}

/**
 * Create or update user registration
 */
export async function saveUserRegistration(registration: UserRegistration): Promise<void> {
  type UserRegistrationInsert = Database['public']['Tables']['user_registrations']['Insert']
  type UserRegistrationUpdate = Database['public']['Tables']['user_registrations']['Update']

  const insertData: UserRegistrationInsert = {
    user_id: registration.userId,
    email: registration.email.toLowerCase(),
    name: registration.name,
    providers: registration.providers,
    registered_at: registration.registeredAt,
  }

  // Try to insert first
  const { error: insertError } = await supabase
    .from('user_registrations')
    // @ts-expect-error - Supabase client with placeholder values has type issues during build
    .insert(insertData)

  // If user already exists (unique constraint violation), update instead
  if (insertError && insertError.code === '23505') {
    const updateData: UserRegistrationUpdate = {
      name: registration.name,
      providers: registration.providers,
    }

    const { error: updateError } = await supabase
      .from('user_registrations')
      // @ts-expect-error - Supabase client with placeholder values has type issues during build
      .update(updateData)
      .eq('user_id', registration.userId)

    if (updateError) {
      console.error('Error updating user registration:', updateError)
      throw new Error(`Failed to update user registration: ${updateError.message}`)
    }
  } else if (insertError) {
    console.error('Error saving user registration:', insertError)
    throw new Error(`Failed to save user registration: ${insertError.message}`)
  }
}
