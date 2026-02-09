-- Add 'oci' to the cloud_provider check constraint
ALTER TABLE installations
  DROP CONSTRAINT IF EXISTS installations_cloud_provider_check;

ALTER TABLE installations
  ADD CONSTRAINT installations_cloud_provider_check
  CHECK (cloud_provider IN ('aws', 'gcp', 'azure', 'oci'));
