-- Atomic subscription items sync: deletes old items and inserts new ones in one transaction
CREATE OR REPLACE FUNCTION sync_subscription_items(
  p_org_id uuid,
  p_items jsonb
) RETURNS void
LANGUAGE plpgsql
SET search_path = ''
AS $$
BEGIN
  DELETE FROM public.subscription_items WHERE org_id = p_org_id;

  INSERT INTO public.subscription_items (org_id, installation_id, stripe_item_id, stripe_price_id, tier, product_name, quantity)
  SELECT
    p_org_id,
    (item->>'installation_id')::uuid,
    item->>'stripe_item_id',
    item->>'stripe_price_id',
    (item->>'tier')::integer,
    item->>'product_name',
    (item->>'quantity')::integer
  FROM jsonb_array_elements(p_items) AS item;
END;
$$;
