-- Migration: Add name and description to installations table
-- Created: 2025-01-25

-- Add name and description columns
ALTER TABLE installations
ADD COLUMN IF NOT EXISTS name TEXT,
ADD COLUMN IF NOT EXISTS description TEXT;

-- Update existing installations to have a default name
UPDATE installations
SET name = COALESCE(subdomain, 'Installation ' || SUBSTRING(installation_key FROM 1 FOR 8))
WHERE name IS NULL;

-- Make name NOT NULL after setting defaults
ALTER TABLE installations
ALTER COLUMN name SET NOT NULL;
