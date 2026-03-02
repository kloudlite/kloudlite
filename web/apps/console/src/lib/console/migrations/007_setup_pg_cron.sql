-- 007_setup_pg_cron.sql
-- Enable pg_cron and pg_net for scheduled Edge Function invocation.
--
-- PREREQUISITES:
--   1. Enable pg_cron extension in Supabase Dashboard → Database → Extensions
--   2. Enable pg_net extension in Supabase Dashboard → Database → Extensions
--
-- After enabling extensions, run the cron.schedule() call below in the
-- Supabase SQL Editor with real values for YOUR_PROJECT_REF and YOUR_SERVICE_ROLE_KEY.

CREATE EXTENSION IF NOT EXISTS pg_cron;
CREATE EXTENSION IF NOT EXISTS pg_net;

-- Schedule the billing cron to run every 5 minutes.
-- Replace the placeholders before running:
--   YOUR_PROJECT_REF     → your Supabase project ref (e.g. abcdefghijkl)
--   YOUR_SERVICE_ROLE_KEY → your service_role key (from Settings → API)
--
-- SELECT cron.schedule(
--   'billing-cron',
--   '*/5 * * * *',
--   $$
--   SELECT net.http_post(
--     url := 'https://YOUR_PROJECT_REF.supabase.co/functions/v1/billing-cron',
--     headers := jsonb_build_object(
--       'Content-Type', 'application/json',
--       'Authorization', 'Bearer YOUR_SERVICE_ROLE_KEY'
--     ),
--     body := '{}'::jsonb
--   ) AS request_id;
--   $$
-- );
