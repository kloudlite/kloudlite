-- Enable Row Level Security on all public tables
-- This migration fixes Supabase linter warnings about RLS being disabled

-- 1. Enable RLS on magic_link_tokens
ALTER TABLE magic_link_tokens ENABLE ROW LEVEL SECURITY;

-- Policy: Only allow backend to manage magic link tokens (no direct user access)
CREATE POLICY "Service role only for magic_link_tokens"
  ON magic_link_tokens
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- 2. Enable RLS on user_registrations
ALTER TABLE user_registrations ENABLE ROW LEVEL SECURITY;

-- Policy: Users can only read their own registration
CREATE POLICY "Users can read own registration"
  ON user_registrations
  FOR SELECT
  TO authenticated
  USING (auth.jwt() ->> 'email' = email);

-- Policy: Service role can manage all registrations
CREATE POLICY "Service role can manage user_registrations"
  ON user_registrations
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- 3. Enable RLS on installations
ALTER TABLE installations ENABLE ROW LEVEL SECURITY;

-- Policy: Users can read installations they own
CREATE POLICY "Users can read own installations"
  ON installations
  FOR SELECT
  TO authenticated
  USING (
    owner_email = auth.jwt() ->> 'email'
    OR id IN (
      SELECT installation_id
      FROM installation_members
      WHERE user_email = auth.jwt() ->> 'email'
    )
  );

-- Policy: Service role can manage all installations
CREATE POLICY "Service role can manage installations"
  ON installations
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- 4. Enable RLS on installation_members
ALTER TABLE installation_members ENABLE ROW LEVEL SECURITY;

-- Policy: Users can read members of installations they belong to
CREATE POLICY "Users can read installation members"
  ON installation_members
  FOR SELECT
  TO authenticated
  USING (
    installation_id IN (
      SELECT id FROM installations WHERE owner_email = auth.jwt() ->> 'email'
    )
    OR installation_id IN (
      SELECT installation_id FROM installation_members WHERE user_email = auth.jwt() ->> 'email'
    )
  );

-- Policy: Service role can manage all members
CREATE POLICY "Service role can manage installation_members"
  ON installation_members
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- 5. Enable RLS on installation_invitations
ALTER TABLE installation_invitations ENABLE ROW LEVEL SECURITY;

-- Policy: Users can read invitations sent to them or from their installations
CREATE POLICY "Users can read installation invitations"
  ON installation_invitations
  FOR SELECT
  TO authenticated
  USING (
    email = auth.jwt() ->> 'email'
    OR installation_id IN (
      SELECT id FROM installations WHERE owner_email = auth.jwt() ->> 'email'
    )
  );

-- Policy: Service role can manage all invitations
CREATE POLICY "Service role can manage installation_invitations"
  ON installation_invitations
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- 6. Enable RLS on domain_reservations
ALTER TABLE domain_reservations ENABLE ROW LEVEL SECURITY;

-- Policy: Service role only (system table)
CREATE POLICY "Service role can manage domain_reservations"
  ON domain_reservations
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);

-- 7. Enable RLS on ip_records
ALTER TABLE ip_records ENABLE ROW LEVEL SECURITY;

-- Policy: Service role only (system table)
CREATE POLICY "Service role can manage ip_records"
  ON ip_records
  FOR ALL
  TO service_role
  USING (true)
  WITH CHECK (true);
