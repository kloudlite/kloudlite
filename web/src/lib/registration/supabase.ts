/**
 * Supabase Client for Registration System
 *
 * Lazy initialization to avoid build-time errors when env vars are not set
 */

import { createClient } from '@supabase/supabase-js'
import type { Database } from './supabase-types'

// Use dummy values for build time - will be replaced at runtime
const supabaseUrl = process.env.SUPABASE_URL || 'https://placeholder.supabase.co'
const supabaseKey = process.env.SUPABASE_KEY || 'placeholder-key'

// Create client with potentially placeholder values
export const supabase = createClient<Database>(supabaseUrl, supabaseKey, {
  auth: {
    persistSession: false, // We're using NextAuth for session management
  },
})

// Re-export Database type for convenience
export type { Database } from './supabase-types'
