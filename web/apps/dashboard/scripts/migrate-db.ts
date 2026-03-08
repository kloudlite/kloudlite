#!/usr/bin/env tsx
/**
 * Database Migration Script (Dashboard)
 *
 * Dashboard shares the same Supabase databases as Console.
 * This script delegates to the console migration.
 * Run `bun run --cwd ../console scripts/migrate-db.ts` instead,
 * or use this wrapper which simply re-exports for convenience.
 */

import { readFileSync } from 'fs'
import { join } from 'path'
import { fileURLToPath } from 'url'
import { Pool } from 'pg'
import dotenv from 'dotenv'

const __filename = fileURLToPath(import.meta.url)
const __dirname = join(__filename, '..')

async function migrate() {
  const databaseUrl = process.env.DATABASE_URL

  if (!databaseUrl) {
    console.error('❌ DATABASE_URL is required.')
    console.error('   Get it from: Supabase Dashboard > Project Settings > Database > Connection String')
    process.exit(1)
  }

  console.log('ℹ️  Dashboard shares the console database.')
  console.log('   Run the console migration script for full schema management:')
  console.log('   bun run --cwd web/apps/console scripts/migrate-db.ts')
  console.log('')
  console.log('   Verifying connection to main database...')

  const pool = new Pool({ connectionString: databaseUrl, ssl: { rejectUnauthorized: false } })

  try {
    await pool.query('SELECT NOW()')
    console.log('✅ Database connection OK')

    const tables = [
      'installations',
      'dns_configurations',
      'domain_reservations',
      'installation_members',
      'installation_invitations',
      'billing_accounts',
    ]

    for (const table of tables) {
      const result = await pool.query(
        `SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`,
        [table],
      )
      const exists = result.rows[0].count > 0
      if (exists) {
        const countResult = await pool.query(`SELECT COUNT(*) FROM ${table}`)
        console.log(`   ✅ ${table} (${countResult.rows[0].count} rows)`)
      } else {
        console.error(`   ❌ ${table} not found — run console migration first`)
      }
    }
  } finally {
    await pool.end()
  }
}

const isMainModule = import.meta.url === `file://${process.argv[1]}`

if (isMainModule) {
  dotenv.config({ path: join(__dirname, '..', '.env.local') })
  dotenv.config({ path: join(__dirname, '..', '.env') })

  migrate()
    .then(() => process.exit(0))
    .catch((error) => {
      console.error('❌ Migration check failed:', error)
      process.exit(1)
    })
}

export { migrate }
