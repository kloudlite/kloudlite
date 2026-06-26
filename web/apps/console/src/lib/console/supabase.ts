import { createRequire } from 'module'
import type { SupabaseClient } from '@supabase/supabase-js'
import type { Database } from './supabase-types'
import type { PiiDatabase } from './supabase-pii-types'

const require = createRequire(import.meta.url)

if (typeof globalThis.WebSocket === 'undefined') {
  // @ts-expect-error
  globalThis.WebSocket = class WebSocketStub {}
}

type SupabaseModule = typeof import('@supabase/supabase-js')
type AnyDatabase = Database | PiiDatabase

const clients = new Map<string, SupabaseClient<AnyDatabase>>()

function createClient(url: string, key: string): SupabaseClient<AnyDatabase> {
  const cacheKey = `${url}::${key}`
  if (clients.has(cacheKey)) return clients.get(cacheKey)! as SupabaseClient<AnyDatabase>
  const { createClient: cc } = require('@supabase/supabase-js') as SupabaseModule
  const client = cc(url, key, { auth: { persistSession: false } }) as SupabaseClient<AnyDatabase>
  clients.set(cacheKey, client)
  return client
}

function createProxy(url: string, key: string) {
  return new Proxy({} as SupabaseClient<AnyDatabase>, {
    get(_, prop) {
      const client = createClient(url, key) as unknown as Record<PropertyKey, unknown>
      const value = client[prop]
      return typeof value === 'function' ? (value as (...args: unknown[]) => unknown).bind(client) : value
    },
  })
}

const supabaseUrl = () => process.env.SUPABASE_URL || process.env.NEXT_PUBLIC_SUPABASE_URL || ''
const supabaseKey = () => process.env.SUPABASE_KEY || process.env.SUPABASE_SERVICE_ROLE_KEY || ''
const piiUrl = () => process.env.PII_SUPABASE_URL || ''
const piiKey = () => process.env.PII_SUPABASE_KEY || ''

export const supabase = createProxy(supabaseUrl(), supabaseKey()) as SupabaseClient<Database>
export const piiSupabase = createProxy(piiUrl(), piiKey()) as SupabaseClient<PiiDatabase>
