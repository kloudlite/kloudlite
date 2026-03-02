-- Migration: Fix function search path security issue
-- Created: 2026-03-02
-- Description: Replace update_updated_at_column() function with one that has fixed search_path
--
-- Security: Fixes "function_search_path_mutable" linter warning by setting search_path

-- ============================================================================
-- 1. REPLACE FUNCTION WITH FIXED SEARCH PATH
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER
SET search_path TO public
LANGUAGE plpgsql
AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;

-- ============================================================================
-- 2. VERIFY FUNCTION WAS UPDATED
-- ============================================================================

SELECT
  proname as function_name,
  prokind as function_type,
  prosecdef as has_fixed_search_path
FROM pg_proc
WHERE proname = 'update_updated_at_column'
  AND pronamespace = 'public'::regnamespace;

-- ============================================================================
-- MIGRATION COMPLETE
-- ============================================================================
