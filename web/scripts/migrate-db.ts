#!/usr/bin/env tsx
/**
 * Database Migration Script
 * Applies schema.sql to Supabase database using PostgreSQL connection
 */

import { readFileSync } from 'fs'
import { join } from 'path'
import { Pool } from 'pg'

async function migrate() {
  // Parse Supabase URL to get database connection info
  const supabaseUrl = process.env.NEXT_PUBLIC_SUPABASE_URL
  const databaseUrl = process.env.DATABASE_URL

  if (!databaseUrl && !supabaseUrl) {
    console.error('❌ Missing required environment variables:')
    console.error('   Either provide:')
    console.error('   - DATABASE_URL (PostgreSQL connection string)')
    console.error('   Or:')
    console.error('   - NEXT_PUBLIC_SUPABASE_URL')
    console.error('')
    console.error('💡 To get your DATABASE_URL from Supabase:')
    console.error('   1. Go to Project Settings > Database')
    console.error('   2. Copy the Connection String (URI format)')
    console.error('   3. Replace [YOUR-PASSWORD] with your database password')
    console.error('   4. Add to .env.local:')
    console.error('      DATABASE_URL="postgresql://postgres:[PASSWORD]@..."')
    process.exit(1)
  }

  let connectionString: string

  if (databaseUrl) {
    connectionString = databaseUrl
  } else {
    // Try to construct from Supabase URL (won't work without password)
    console.error('❌ DATABASE_URL is required. Please add it to your .env.local file.')
    console.error('   Get it from: Supabase Dashboard > Project Settings > Database > Connection String')
    process.exit(1)
  }

  console.log('🔌 Connecting to PostgreSQL database...')

  const pool = new Pool({
    connectionString,
    ssl: {
      rejectUnauthorized: false, // Required for Supabase
    },
  })

  try {
    // Test connection
    await pool.query('SELECT NOW()')
    console.log('✅ Database connection established')

    // Read schema.sql
    const schemaPath = join(__dirname, '..', 'src', 'lib', 'registration', 'schema.sql')
    console.log(`📄 Reading schema from: ${schemaPath}`)

    const sql = readFileSync(schemaPath, 'utf-8')

    console.log('')
    console.log('🚀 Executing migration...')
    console.log('   This will:')
    console.log('   - Drop existing tables (if any)')
    console.log('   - Create new tables with updated schema')
    console.log('   - Set up triggers and indexes')
    console.log('')

    // Execute the migration
    await pool.query(sql)

    console.log('✅ Migration SQL executed successfully!')

    // Verify tables were created
    console.log('')
    console.log('🔍 Verifying tables...')

    const tables = [
      'user_registrations',
      'installations',
      'ip_records',
      'domain_reservations',
      'tls_certificates',
    ]

    for (const table of tables) {
      const result = await pool.query(
        `SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1`,
        [table]
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

    console.log('')
    console.log('✨ Database migration complete!')
  } catch (error) {
    console.error('')
    console.error('❌ Migration failed:')
    if (error instanceof Error) {
      console.error(`   ${error.message}`)
      if (error.stack) {
        console.error('')
        console.error('Stack trace:')
        console.error(error.stack)
      }
    } else {
      console.error(error)
    }
    process.exit(1)
  } finally {
    await pool.end()
  }
}

// Handle command line execution
if (require.main === module) {
  // Load environment variables from .env.local
  require('dotenv').config({ path: join(__dirname, '..', '.env.local') })
  require('dotenv').config({ path: join(__dirname, '..', '.env') })

  migrate()
    .then(() => process.exit(0))
    .catch((error) => {
      console.error(error)
      process.exit(1)
    })
}

export { migrate }
