-- Add edge_certificate_pack_id column to installations table
-- This column stores the CloudFlare Edge Certificate Pack ID for wildcard subdomain support

ALTER TABLE installations
ADD COLUMN IF NOT EXISTS edge_certificate_pack_id TEXT;

-- Add comment to explain the column purpose
COMMENT ON COLUMN installations.edge_certificate_pack_id IS
'CloudFlare Edge Certificate Pack ID for wildcard subdomain (*.subdomain.khost.dev) support';
