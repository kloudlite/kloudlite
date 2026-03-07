# Stripe Billing Migration Design

## Context

Kloudlite currently uses Razorpay Orders (one-time payments) with a custom subscription lifecycle — manual renewals via Supabase cron, custom proration logic, and an in-page checkout popup. This design replaces the entire billing system with Stripe-native subscriptions, eliminating ~2000 lines of custom billing code.

## Decisions

| Aspect | Decision |
|---|---|
| Payment provider | Stripe (replacing Razorpay) |
| Subscription management | Stripe-native subscriptions |
| Pricing | Flat control plane ($29/mo) + per-seat tiers ($29/$49/$89), all prepaid, no metered/overage |
| Currency | Multi-currency (configured in Stripe Dashboard) |
| Checkout | Stripe Checkout (hosted page redirect) |
| Self-service | Stripe Customer Portal (invoices, payment methods, cancellation) |
| Plan changes | In-app UI → server action → Stripe API, proration by Stripe (`always_invoice`) |
| Renewals | Fully automatic via Stripe |
| Data migration | Fresh start — no existing customers to migrate |

## Stripe Products & Prices

Configured in Stripe Dashboard (not in code):

- **Control Plane** — $29/mo flat recurring
- **Tier 1 – Light Workloads** — $29/user/mo per-seat recurring
- **Tier 2 – Standard Workloads** — $49/user/mo per-seat recurring
- **Tier 3 – Power Users** — $89/user/mo per-seat recurring

No metered prices. No overage billing. Users get fixed resources per seat; to get more, they upgrade seats or tier.

One Stripe Customer per installation. One Stripe Subscription per installation with multiple line items (control plane + selected tier seats).

## Database Schema

Two tables replace the current six (`subscription_plans`, `subscriptions`, `invoices`, `renewal_jobs`, `cron_job_logs`, `processed_webhook_events`):

### `stripe_customers`

| Column | Type | Purpose |
|---|---|---|
| `id` | UUID PK | |
| `installation_id` | FK installations, UNIQUE | |
| `stripe_customer_id` | TEXT UNIQUE | Stripe Customer ID |
| `stripe_subscription_id` | TEXT UNIQUE | Stripe Subscription ID |
| `billing_status` | TEXT | `active`, `past_due`, `cancelled`, `trialing` |
| `payment_issue` | BOOLEAN DEFAULT false | Show warning banner when true |
| `current_period_end` | TIMESTAMPTZ | When current billing cycle ends |
| `created_at` | TIMESTAMPTZ DEFAULT now() | |
| `updated_at` | TIMESTAMPTZ DEFAULT now() | |

### `subscription_items` (entitlements cache)

| Column | Type | Purpose |
|---|---|---|
| `id` | UUID PK | |
| `installation_id` | FK installations | |
| `stripe_subscription_item_id` | TEXT UNIQUE | Needed for Stripe API calls |
| `stripe_price_id` | TEXT | Which Stripe price |
| `tier` | INT | 0 (control plane), 1, 2, or 3 |
| `product_name` | TEXT | "Control Plane", "Tier 1 – Light", etc. |
| `quantity` | INT | Number of seats (1 for control plane) |
| `created_at` | TIMESTAMPTZ DEFAULT now() | |
| `updated_at` | TIMESTAMPTZ DEFAULT now() | |

### `stripe_webhook_events` (idempotency guard)

| Column | Type | Purpose |
|---|---|---|
| `id` | UUID PK | |
| `stripe_event_id` | TEXT UNIQUE | Stripe event ID |
| `event_type` | TEXT | Event type string |
| `processed_at` | TIMESTAMPTZ DEFAULT now() | |

### Tables dropped

- `subscription_plans` — plans defined in Stripe Dashboard
- `subscriptions` — managed by Stripe
- `invoices` — managed by Stripe, viewable in Customer Portal
- `renewal_jobs` — Stripe auto-renews
- `cron_job_logs` — no cron needed
- `processed_webhook_events` — replaced by `stripe_webhook_events`
- `pg_cron` schedule — no cron needed

## User Flows

### Flow 1: New Subscription

1. User creates installation, picks tier + seat count in app UI
2. Server action creates a Stripe Customer (stores `installation_id` in metadata)
3. Server action creates a Stripe Checkout Session with line items (control plane + tier seats)
4. User is redirected to Stripe Checkout hosted page
5. User pays on Stripe's page
6. Stripe redirects back to `/installations/[id]?checkout=success`
7. Stripe webhook `checkout.session.completed` fires
8. Webhook handler upserts `stripe_customers`, syncs `subscription_items`, activates installation

### Flow 2: Modify Subscription (Add/Remove Tiers or Seats)

1. User changes seat counts or adds a new tier in app's plan configurator
2. Server action calls `stripe.subscriptions.update()` with new items and `proration_behavior: 'always_invoice'`
3. Stripe calculates proration, invoices the difference immediately
4. Stripe webhook `customer.subscription.updated` fires
5. Webhook handler syncs `subscription_items` (upsert new, delete removed, update quantities)

### Flow 3: Cancel Subscription

1. User clicks "Manage Billing" → opens Stripe Customer Portal
2. User cancels in Stripe Portal
3. Stripe sets `cancel_at_period_end = true`
4. Webhook `customer.subscription.updated` fires — update `billing_status`
5. At period end, webhook `customer.subscription.deleted` fires
6. Webhook handler deactivates installation

### Flow 4: Payment Failure / Renewal

Stripe auto-charges at period end. On failure, Stripe retries via Smart Retries and sends dunning emails. If all retries fail, subscription is deleted and webhook deactivates the installation. No custom code needed.

## Webhook Handler

`POST /api/stripe/webhook` handles these events:

| Event | Action |
|---|---|
| `checkout.session.completed` | Upsert `stripe_customers`, sync `subscription_items`, activate installation |
| `customer.subscription.updated` | Sync `billing_status`, `current_period_end`, `subscription_items`, `payment_issue` flag |
| `customer.subscription.deleted` | Set `billing_status = 'cancelled'`, deactivate installation |
| `invoice.payment_failed` | Set `payment_issue = true` on `stripe_customers` |

Idempotency enforced via `stripe_webhook_events` table.

## Billing UI

Simplified to: status display + seat modifier + "Manage Billing" button.

```
┌─────────────────────────────────────────────────────────┐
│  Billing                                [Manage Billing ↗]  │
│                                                              │
│  Status: ● Active    Next billing: Apr 8, 2026               │
│                                                              │
│  Current Products                                            │
│    Control Plane          1×    $29/mo                        │
│    Tier 1 – Light         2×    $58/mo                        │
│    Tier 2 – Standard      3×    $147/mo                       │
│    Total: $234/mo                                            │
│                                                              │
│  Modify Plan                                                 │
│    Tier 1 – Light    [−] 2 [+]  $29/user/mo                  │
│    Tier 2 – Standard [−] 3 [+]  $49/user/mo                  │
│    Tier 3 – Power    [−] 0 [+]  $89/user/mo                  │
│    [Save Changes]                                            │
│    Prorated charges applied automatically by Stripe.         │
└─────────────────────────────────────────────────────────────┘
```

### Components kept (simplified):
- `subscription-management.tsx` — orchestrator: status + configurator + "Manage Billing"
- `subscription-configurator.tsx` — tier picker + seat count
- `subscription-status.tsx` — active products, quantities, costs, next billing date
- `subscription-header.tsx` — status badge

### Components created:
- `payment-warning-banner.tsx` — shown when `payment_issue = true`, links to Stripe Portal

### Components deleted:
- `payment-due-banner.tsx` — no manual renewal
- `invoice-history.tsx` — Stripe Portal shows invoices
- `past-subscriptions.tsx` — Stripe Portal shows history
- `razorpay-provider.tsx` — no popup checkout

## File Changes Summary

### Delete (~2500 lines)
- `src/lib/razorpay.ts`, `src/lib/razorpay-types.ts`
- `src/components/razorpay-provider.tsx`
- `src/app/actions/billing/proration.ts`, `src/app/actions/billing/verification.ts`
- `src/app/api/razorpay/webhook/route.ts`
- `src/components/billing/payment-due-banner.tsx`
- `src/components/billing/invoice-history.tsx`
- `src/components/billing/past-subscriptions.tsx`
- `src/app/actions/billing/proration.test.ts`
- `supabase/functions/billing-cron/index.ts`

### Create (~500 lines)
- `src/lib/stripe.ts`
- `src/app/api/stripe/webhook/route.ts`
- `src/app/actions/billing/checkout.ts`
- `src/components/billing/payment-warning-banner.tsx`
- DB migration `014_stripe_migration.sql`

### Modify
- `src/app/actions/billing/queries.ts`, `src/app/actions/billing/subscriptions.ts`
- `src/lib/console/storage/billing.ts`, `src/lib/console/storage/billing-types.ts`
- `src/lib/console/supabase-types.ts`, `src/lib/console/storage/index.ts`
- `src/components/billing/subscription-management.tsx`
- `src/components/billing/subscription-configurator.tsx`
- `src/components/billing/subscription-header.tsx`, `src/components/billing/subscription-status.tsx`
- `src/hooks/use-subscription-payments.ts`
- `src/components/kl-cloud-installation-form.tsx`
- `src/app/installations/[id]/billing/page.tsx`
- `src/proxy.ts`, `.env.example`, `package.json`
- `.github/workflows/deploy-edge-functions.yml`
- `src/lib/billing-utils.ts`, `src/lib/billing-utils.test.ts`

### Net impact
- ~2000 lines deleted
- Custom proration math, renewal cron, manual payment flow all eliminated
- Stripe handles renewals, proration, dunning, retries, invoice generation
