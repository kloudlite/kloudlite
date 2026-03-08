-- Atomic org creation: creates org + owner member in one transaction
CREATE OR REPLACE FUNCTION create_organization_with_owner(
  p_name text, p_slug text, p_created_by text
) RETURNS uuid
LANGUAGE plpgsql
SET search_path = ''
AS $$
DECLARE
  v_org_id uuid;
BEGIN
  INSERT INTO public.organizations (name, slug, created_by)
  VALUES (p_name, p_slug, p_created_by)
  RETURNING id INTO v_org_id;

  INSERT INTO public.organization_members (org_id, user_id, role, added_by)
  VALUES (v_org_id, p_created_by, 'owner', p_created_by);

  RETURN v_org_id;
END;
$$;
