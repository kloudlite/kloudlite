/**
 * Supabase Client for PII Database
 *
 * Separate database instance for personally identifiable information
 * (users, magic link tokens, contact messages)
 */

import { createRequire } from 'module'
import type { SupabaseClient } from '@supabase/supabase-js'
import type { PiiDatabase } from './supabase-pii-types'

const require = createRequire(import.meta.url)

if (typeof globalThis.WebSocket === 'undefined') {
  // @ts-expect-error - assigning stub for non-browser runtime compatibility
  globalThis.WebSocket = class WebSocketStub {}
}

type SupabaseModule = typeof import('@supabase/supabase-js')
let cachedClient: SupabaseClient<PiiDatabase> | null = null

function getPiiSupabaseClient(): SupabaseClient<PiiDatabase> {
  if (cachedClient) return cachedClient

  const supabaseUrl = process.env.PII_SUPABASE_URL || 'https://placeholder.supabase.co'
  const supabaseKey = process.env.PII_SUPABASE_KEY || 'placeholder-key'
  const { createClient } = require('@supabase/supabase-js') as SupabaseModule

  cachedClient = createClient<PiiDatabase>(supabaseUrl, supabaseKey, {
    auth: {
      persistSession: false,
    },
  }) as SupabaseClient<PiiDatabase>

  return cachedClient
}

export const piiSupabase = new Proxy({} as SupabaseClient<PiiDatabase>, {
  get(_target, prop) {
    const client = getPiiSupabaseClient() as unknown as Record<PropertyKey, unknown>
    const value = client[prop]
    return typeof value === 'function' ? (value as (...args: unknown[]) => unknown).bind(client) : value
  },
}) as SupabaseClient<PiiDatabase>

export type { PiiDatabase } from './supabase-pii-types'
