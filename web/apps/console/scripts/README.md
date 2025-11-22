# Database Migration

This directory contains scripts for managing the Supabase database schema.

## Running the Migration

The database schema needs to be applied to your Supabase database before the application can work properly.

### Prerequisites

You need to add your Supabase database connection string to `.env.local`:

1. Go to your Supabase Dashboard
2. Navigate to **Project Settings** > **Database**
3. Copy the **Connection String** (URI format)
4. Replace `[YOUR-PASSWORD]` with your actual database password
5. Add to `.env.local`:

```env
DATABASE_URL="postgresql://postgres:[YOUR-PASSWORD]@db.[project-ref].supabase.co:5432/postgres"
```

### Run Migration

```bash
pnpm db:migrate
```

This will:

- Drop existing tables (if any)
- Create new tables with the updated schema:
  - `user_registrations` - User authentication data
  - `installations` - Installation-specific data (multiple per user)
  - `ip_records` - IP addresses for installations and workmachines
  - `domain_reservations` - Subdomain reservations
  - `tls_certificates` - TLS certificates for installations
- Set up triggers and indexes
- Verify all tables were created successfully

## Troubleshooting

### Missing DATABASE_URL

If you see this error:

```
❌ DATABASE_URL is required. Please add it to your .env.local file.
```

Follow the prerequisites section above to add your database connection string.

### Connection Failed

If the migration fails to connect:

- Verify your database password is correct
- Check that your IP is allowed in Supabase (Project Settings > Database > Connection Pooling > Allow connections from)
- Ensure you're using the direct database connection string, not the pooler connection string

### Tables Already Exist

The migration script will automatically drop existing tables before creating new ones. This is safe during development but **be careful in production** as it will delete all data.

## Manual Migration (Alternative)

If you prefer not to use the script, you can manually run the migration:

1. Go to your Supabase Dashboard
2. Navigate to **SQL Editor**
3. Copy the contents of `src/lib/registration/schema.sql`
4. Paste and execute in the SQL Editor

This achieves the same result without needing to configure DATABASE_URL locally.
