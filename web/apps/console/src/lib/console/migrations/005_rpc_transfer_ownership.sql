-- Atomic ownership transfer: demotes old owner + promotes new owner in one transaction
CREATE OR REPLACE FUNCTION transfer_org_ownership(
  p_org_id uuid, p_old_owner text, p_new_owner text
) RETURNS void
LANGUAGE plpgsql
SET search_path = ''
AS $$
BEGIN
  UPDATE public.organization_members
  SET role = 'admin', updated_at = now()
  WHERE org_id = p_org_id AND user_id = p_old_owner AND role = 'owner';

  IF NOT FOUND THEN
    RAISE EXCEPTION 'Current owner not found';
  END IF;

  UPDATE public.organization_members
  SET role = 'owner', updated_at = now()
  WHERE org_id = p_org_id AND user_id = p_new_owner;

  IF NOT FOUND THEN
    RAISE EXCEPTION 'New owner not found in organization';
  END IF;
END;
$$;
