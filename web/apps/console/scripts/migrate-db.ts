#!/usr/bin/env tsx
/**
 * Database Migration Script
 * Applies schema migrations to both Main and PII Supabase databases
 */

import { readFileSync } from 'fs'
import { join } from 'path'
import { fileURLToPath } from 'url'
import { Pool } from 'pg'
import dotenv from 'dotenv'

const __filename = fileURLToPath(import.meta.url)
const __dirname = join(__filename, '..')

async function runMigration(pool: Pool, schemaPath: string, label: string, tables: string[]) {
  console.log(`\n📄 [${label}] Reading schema from: ${schemaPath}`)
  const sql = readFileSync(schemaPath, 'utf-8')

  console.log(`🚀 [${label}] Executing migration...`)
  await pool.query(sql)
  console.log(`✅ [${label}] Migration SQL executed successfully!`)

  console.log(`🔍 [${label}] Verifying tables...`)
  for (const table of tables) {
    const result = await pool.query(
      `SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`,
      [table],
    )
    const exists = result.rows[0].count > 0
    if (exists) {
      const countResult = await pool.query(`SELECT COUNT(*) FROM ${table}`)
      const rowCount = countResult.rows[0].count
      console.log(`   ✅ ${table} (${rowCount} rows)`)
    } else {
      console.error(`   ❌ ${table} not found`)
    }
  }
}

async function migrate() {
  const databaseUrl = process.env.DATABASE_URL
  const piiDatabaseUrl = process.env.PII_DATABASE_URL

  if (!databaseUrl) {
    console.error('❌ DATABASE_URL is required for the main database.')
    console.error('   Get it from: Supabase Dashboard > Project Settings > Database > Connection String')
    process.exit(1)
  }

  const migrationsDir = join(__dirname, '..', 'src', 'lib', 'console', 'migrations')

  // --- Main DB ---
  console.log('🔌 Connecting to Main database...')
  const mainPool = new Pool({ connectionString: databaseUrl, ssl: { rejectUnauthorized: false } })

  try {
    await mainPool.query('SELECT NOW()')
    console.log('✅ Main database connection established')

    await runMigration(mainPool, join(migrationsDir, '001_schema.sql'), 'Main', [
      'organizations',
      'organization_members',
      'organization_invitations',
      'installations',
      'dns_configurations',
      'domain_reservations',
      'billing_accounts',
      'subscription_items',
      'processed_webhook_events',
    ])
  } finally {
    await mainPool.end()
  }

  // --- PII DB ---
  if (piiDatabaseUrl) {
    console.log('\n🔌 Connecting to PII database...')
    const piiPool = new Pool({ connectionString: piiDatabaseUrl, ssl: { rejectUnauthorized: false } })

    try {
      await piiPool.query('SELECT NOW()')
      console.log('✅ PII database connection established')

      await runMigration(piiPool, join(migrationsDir, '001_pii_schema.sql'), 'PII', [
        'users',
        'magic_link_tokens',
        'contact_messages',
      ])
    } finally {
      await piiPool.end()
    }
  } else {
    console.log('\n⚠️  PII_DATABASE_URL not set — skipping PII database migration.')
    console.log('   Set PII_DATABASE_URL in .env.local to migrate the PII database.')
  }

  console.log('\n✨ Database migration complete!')
}

const isMainModule = import.meta.url === `file://${process.argv[1]}`

if (isMainModule) {
  dotenv.config({ path: join(__dirname, '..', '.env.local') })
  dotenv.config({ path: join(__dirname, '..', '.env') })

  migrate()
    .then(() => process.exit(0))
    .catch((error) => {
      console.error('❌ Migration failed:', error)
      process.exit(1)
    })
}

export { migrate }
