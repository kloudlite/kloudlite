# Database Migration Instructions

## Migration: Unified DNS Architecture Schema Updates (2025-11-04)

This migration updates the database schema to support the unified DNS architecture by:
1. Adding origin certificate fields to the installations table
2. Migrating ip_records from type-based to domain-based schema
3. Simplifying edge_certificates by removing the scope concept

## Prerequisites

- Access to your Supabase project
- Supabase CLI installed (optional but recommended)
- Database backup (highly recommended before running migrations)

## Option 1: Execute via Supabase Dashboard (Recommended)

1. **Backup your database first!**
   - Go to Supabase Dashboard → Database → Backups
   - Create a backup before proceeding

2. **Open SQL Editor**
   - Navigate to: Supabase Dashboard → SQL Editor

3. **Execute the migration**
   - Open the file: `2025-11-04_schema_updates.sql`
   - Copy the entire contents
   - Paste into the SQL Editor
   - Click "Run" to execute

4. **Verify the migration**
   - Check the output for "Migration completed successfully!" message
   - Verify tables have been updated:
     ```sql
     -- Check installations table
     SELECT column_name, data_type
     FROM information_schema.columns
     WHERE table_name = 'installations'
     AND column_name LIKE 'origin_%';

     -- Check ip_records table
     SELECT column_name, data_type
     FROM information_schema.columns
     WHERE table_name = 'ip_records';

     -- Check edge_certificates table
     SELECT column_name, data_type
     FROM information_schema.columns
     WHERE table_name = 'edge_certificates';
     ```

## Option 2: Execute via Supabase CLI

```bash
# Navigate to the migrations directory
cd web/migrations

# Execute the migration
supabase db execute --file 2025-11-04_schema_updates.sql

# Or if you have a connection string
psql $DATABASE_URL -f 2025-11-04_schema_updates.sql
```

## Option 3: Execute via psql

```bash
# Connect to your database
psql "postgresql://user:password@host:port/database"

# Execute the migration
\i /path/to/2025-11-04_schema_updates.sql
```

## Rolling Back

If you need to rollback the migration:

⚠️ **WARNING:** Rollback may result in data loss! The old schema cannot fully represent the new data structure.

```bash
# Via Supabase Dashboard SQL Editor
# Copy contents of 2025-11-04_schema_updates_rollback.sql and execute

# Via CLI
supabase db execute --file 2025-11-04_schema_updates_rollback.sql

# Via psql
psql $DATABASE_URL -f 2025-11-04_schema_updates_rollback.sql
```

## What Changes Were Made

### 1. Installations Table
**Added columns:**
- `origin_certificate` (TEXT) - Cloudflare Origin Certificate PEM
- `origin_private_key` (TEXT) - Private key for the origin certificate
- `origin_cert_id` (TEXT) - Cloudflare certificate ID
- `origin_cert_valid_from` (TIMESTAMPTZ) - Certificate validity start
- `origin_cert_valid_until` (TIMESTAMPTZ) - Certificate validity end

### 2. IP Records Table
**Schema transformation:**

**Old schema:**
- `type` ('installation' | 'workmachine')
- `work_machine_name` (TEXT)
- `dns_record_ids` (TEXT[])

**New schema:**
- `domain_request_name` (TEXT) - Direct domain request identifier
- `ssh_record_id` (TEXT) - Single SSH A record ID
- `route_record_ids` (TEXT[]) - Array of route CNAME record IDs
- `domain_routes` (JSONB) - Array of domain route configurations

**Unique constraint:** `(installation_id, domain_request_name)`

### 3. Edge Certificates Table
**Schema simplification:**

**Old schema:**
- `scope` ('installation' | 'workmachine' | 'domainrequest')
- `scope_identifier` (TEXT)

**New schema:**
- `domain_request_name` (TEXT) - Direct domain request identifier

**Unique constraint:** `(installation_id, domain_request_name)`

## Data Migration Notes

### IP Records Migration
- Only records with `type='installation'` are migrated
- `work_machine_name` → `domain_request_name`
- First DNS record ID → `ssh_record_id`
- Remaining DNS record IDs → `route_record_ids`
- `domain_routes` initialized as empty array `[]`

### Edge Certificates Migration
- Only records with `scope='domainrequest'` are migrated
- `scope_identifier` → `domain_request_name`
- Other scope types are not migrated (assumed deprecated)

## Post-Migration Steps

1. **Regenerate Supabase types** (if using TypeScript):
   ```bash
   npx supabase gen types typescript --local > src/lib/console/supabase-types.ts
   ```

2. **Test the application**:
   - Verify DNS record creation works
   - Test edge certificate generation
   - Check IP record management

3. **Monitor for errors**:
   - Watch application logs
   - Check Supabase logs for any database errors

## Troubleshooting

### Migration fails with constraint violation
- Check if there are duplicate records that violate the new unique constraints
- Manually resolve duplicates before re-running migration

### Data missing after migration
- Check if your data had `type='workmachine'` in ip_records (not migrated by default)
- Check if edge_certificates had `scope='installation'` or `scope='workmachine'` (not migrated)
- Restore from backup and adjust migration logic if needed

### Application errors after migration
- Verify all code changes were deployed (commits: 431ea86b1, 94c92b865)
- Check that Supabase types were regenerated
- Ensure environment variables are correct

## Support

If you encounter issues:
1. Check the migration output for specific error messages
2. Verify your backup is intact before attempting rollback
3. Review the application logs for runtime errors
