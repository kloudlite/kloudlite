/**
 * Supabase Client for Registration System
 *
 * Lazy initialization to avoid build-time errors when env vars are not set
 * Note: Supabase uses Node.js APIs - routes using this must specify runtime: 'nodejs'
 */

import { createRequire } from 'module'
import type { SupabaseClient } from '@supabase/supabase-js'
import type { Database } from './supabase-types'

const require = createRequire(import.meta.url)

// Some build-time evaluation contexts do not provide a WebSocket global.
// Supabase client imports realtime dependencies that expect it to exist.
if (typeof globalThis.WebSocket === 'undefined') {
  // Minimal stub to prevent import-time crashes during build config collection.
  // Realtime features are not used by this app.
  // @ts-expect-error - assigning stub for non-browser runtime compatibility
  globalThis.WebSocket = class WebSocketStub {}
}

type SupabaseModule = typeof import('@supabase/supabase-js')
let cachedClient: SupabaseClient<Database> | null = null

function getSupabaseClient(): SupabaseClient<Database> {
  if (cachedClient) return cachedClient

  const supabaseUrl = process.env.SUPABASE_URL || process.env.NEXT_PUBLIC_SUPABASE_URL || 'https://placeholder.supabase.co'
  const supabaseKey = process.env.SUPABASE_KEY || process.env.SUPABASE_SERVICE_ROLE_KEY || 'placeholder-key'
  const { createClient } = require('@supabase/supabase-js') as SupabaseModule

  cachedClient = createClient<Database>(supabaseUrl, supabaseKey, {
    auth: {
      persistSession: false, // We're using NextAuth for session management
    },
  }) as SupabaseClient<Database>

  return cachedClient
}

export const supabase = new Proxy({} as SupabaseClient<Database>, {
  get(_target, prop) {
    const client = getSupabaseClient() as unknown as Record<PropertyKey, unknown>
    const value = client[prop]
    return typeof value === 'function' ? (value as (...args: unknown[]) => unknown).bind(client) : value
  },
}) as SupabaseClient<Database>

// Re-export Database type for convenience
export type { Database } from './supabase-types'
