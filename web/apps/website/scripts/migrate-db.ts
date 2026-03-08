#!/usr/bin/env tsx
/**
 * Database Migration Script (Website)
 *
 * The website only uses the PII database (for contact_messages).
 * This script verifies the PII DB connection and checks the contact_messages table.
 * Full schema management is handled by the console migration script.
 */

import { join } from 'path'
import { fileURLToPath } from 'url'
import { Pool } from 'pg'
import dotenv from 'dotenv'

const __filename = fileURLToPath(import.meta.url)
const __dirname = join(__filename, '..')

async function migrate() {
  const piiDatabaseUrl = process.env.PII_DATABASE_URL

  if (!piiDatabaseUrl) {
    console.error('❌ PII_DATABASE_URL is required for the website.')
    console.error('   The website uses the PII database for contact_messages.')
    console.error('   Get it from: Supabase Dashboard > Project Settings > Database > Connection String')
    process.exit(1)
  }

  console.log('ℹ️  Website uses the PII database for contact submissions.')
  console.log('   Run the console migration script for full schema management:')
  console.log('   bun run --cwd web/apps/console scripts/migrate-db.ts')
  console.log('')
  console.log('   Verifying connection to PII database...')

  const pool = new Pool({ connectionString: piiDatabaseUrl, ssl: { rejectUnauthorized: false } })

  try {
    await pool.query('SELECT NOW()')
    console.log('✅ PII database connection OK')

    const result = await pool.query(
      `SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'contact_messages'`,
    )
    const exists = result.rows[0].count > 0
    if (exists) {
      const countResult = await pool.query('SELECT COUNT(*) FROM contact_messages')
      console.log(`   ✅ contact_messages (${countResult.rows[0].count} rows)`)
    } else {
      console.error('   ❌ contact_messages not found — run console migration first')
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
