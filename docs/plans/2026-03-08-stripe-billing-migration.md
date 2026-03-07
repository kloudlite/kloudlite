# Stripe Billing Migration — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace Razorpay billing with Stripe-native subscriptions, eliminating ~2000 lines of custom billing code.

**Architecture:** Stripe manages subscription lifecycle, renewals, proration, and dunning. Console app creates Checkout Sessions for new subscriptions, updates subscriptions via Stripe API for plan changes, and syncs state via webhooks into a 2-table local cache (`stripe_customers` + `subscription_items`). Stripe Customer Portal handles payment methods, invoices, and cancellation.

**Tech Stack:** `stripe` npm package, Next.js 16 API routes + Server Actions, Supabase/PostgreSQL, React 19.

**Design doc:** `docs/plans/2026-03-08-stripe-billing-migration-design.md`

---

### Task 1: Swap npm dependency & create Stripe client

**Files:**
- Modify: `web/apps/console/package.json` — remove `razorpay`, add `stripe`
- Create: `web/apps/console/src/lib/stripe.ts`
- Modify: `web/apps/console/.env.example`

**Step 1: Remove razorpay, add stripe**

```bash
cd web/apps/console && bun remove razorpay && bun add stripe
```

**Step 2: Update .env.example**

Replace the Razorpay env vars:
```
RAZORPAY_KEY_ID=your-razorpay-key-id
RAZORPAY_KEY_SECRET=your-razorpay-key-secret
RAZORPAY_WEBHOOK_SECRET=your-razorpay-webhook-secret
NEXT_PUBLIC_RAZORPAY_KEY_ID=your-razorpay-key-id
```

With:
```
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY=pk_test_...
```

**Step 3: Create Stripe singleton**

Create `web/apps/console/src/lib/stripe.ts`:
```typescript
import Stripe from 'stripe'

let instance: Stripe | null = null

export function getStripe(): Stripe {
  if (!instance) {
    const secretKey = process.env.STRIPE_SECRET_KEY
    if (!secretKey) {
      throw new Error('STRIPE_SECRET_KEY must be set')
    }
    instance = new Stripe(secretKey, {
      apiVersion: '2025-04-30.basil',
    })
  }
  return instance
}
```

**Step 4: Delete old Razorpay files**

```bash
rm web/apps/console/src/lib/razorpay.ts
rm web/apps/console/src/lib/razorpay-types.ts
```

**Step 5: Commit**

```bash
git add -A && git commit -m "feat(billing): swap razorpay dep for stripe and create client singleton"
```

---

### Task 2: Database migration

**Files:**
- Create: `web/apps/console/src/lib/console/migrations/014_stripe_migration.sql`

**Step 1: Write migration**

Create `web/apps/console/src/lib/console/migrations/014_stripe_migration.sql`:
```sql
-- Migration: Replace Razorpay billing tables with Stripe tables
-- This is a fresh start — no data migration needed.

-- 1. Drop old billing tables (order matters for FKs)
DROP TABLE IF EXISTS processed_webhook_events CASCADE;
DROP TABLE IF EXISTS cron_job_logs CASCADE;
DROP TABLE IF EXISTS renewal_jobs CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS subscription_plans CASCADE;

-- 2. Remove pg_cron schedule if it exists
DO $$
BEGIN
  PERFORM cron.unschedule('billing-cron-job');
EXCEPTION WHEN OTHERS THEN
  -- pg_cron not installed or job doesn't exist, skip
  NULL;
END $$;

-- 3. Create stripe_customers
CREATE TABLE stripe_customers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  stripe_customer_id TEXT NOT NULL UNIQUE,
  stripe_subscription_id TEXT UNIQUE,
  billing_status TEXT NOT NULL DEFAULT 'active'
    CHECK (billing_status IN ('active', 'past_due', 'cancelled', 'trialing', 'incomplete')),
  payment_issue BOOLEAN NOT NULL DEFAULT false,
  current_period_end TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_stripe_customers_installation UNIQUE (installation_id)
);

-- 4. Create subscription_items (entitlements cache)
CREATE TABLE subscription_items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  stripe_subscription_item_id TEXT NOT NULL UNIQUE,
  stripe_price_id TEXT NOT NULL,
  tier INT NOT NULL CHECK (tier >= 0 AND tier <= 3),
  product_name TEXT NOT NULL,
  quantity INT NOT NULL DEFAULT 1 CHECK (quantity >= 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subscription_items_installation ON subscription_items(installation_id);

-- 5. Create stripe_webhook_events (idempotency)
CREATE TABLE stripe_webhook_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  stripe_event_id TEXT NOT NULL UNIQUE,
  event_type TEXT NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 6. Enable RLS
ALTER TABLE stripe_customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE subscription_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE stripe_webhook_events ENABLE ROW LEVEL SECURITY;

-- RLS: stripe_customers readable by installation members
CREATE POLICY stripe_customers_select ON stripe_customers
  FOR SELECT USING (
    installation_id IN (
      SELECT id FROM installations WHERE user_id = auth.uid()
    )
    OR installation_id IN (
      SELECT installation_id FROM installation_members WHERE user_id = auth.uid()
    )
  );

-- RLS: subscription_items readable by installation members
CREATE POLICY subscription_items_select ON subscription_items
  FOR SELECT USING (
    installation_id IN (
      SELECT id FROM installations WHERE user_id = auth.uid()
    )
    OR installation_id IN (
      SELECT installation_id FROM installation_members WHERE user_id = auth.uid()
    )
  );

-- webhook_events: service_role only (no user policies)
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(billing): add stripe migration — new tables, drop razorpay tables"
```

---

### Task 3: Storage layer — types and data access

**Files:**
- Rewrite: `web/apps/console/src/lib/console/storage/billing-types.ts`
- Rewrite: `web/apps/console/src/lib/console/storage/billing.ts`
- Modify: `web/apps/console/src/lib/console/storage/index.ts`
- Update: `web/apps/console/src/lib/console/supabase-types.ts` — add new table types

**Step 1: Rewrite billing-types.ts**

```typescript
export interface StripeCustomer {
  id: string
  installationId: string
  stripeCustomerId: string
  stripeSubscriptionId: string | null
  billingStatus: 'active' | 'past_due' | 'cancelled' | 'trialing' | 'incomplete'
  paymentIssue: boolean
  currentPeriodEnd: string | null
  createdAt: string
  updatedAt: string
}

export interface SubscriptionItem {
  id: string
  installationId: string
  stripeSubscriptionItemId: string
  stripePriceId: string
  tier: number
  productName: string
  quantity: number
  createdAt: string
  updatedAt: string
}
```

**Step 2: Rewrite billing.ts**

Replace all 547 lines with ~150 lines covering:
- `getStripeCustomer(installationId)` — get Stripe customer mapping
- `upsertStripeCustomer(data)` — create or update Stripe customer mapping
- `updateBillingStatus(installationId, status, currentPeriodEnd?, paymentIssue?)` — update status from webhook
- `getSubscriptionItems(installationId)` — get current entitlements
- `syncSubscriptionItems(installationId, items[])` — full sync from webhook (upsert matching, delete removed)
- `getActiveSubscriptionsByInstallationIds(ids)` — batch lookup for installation list pages
- `isWebhookEventProcessed(eventId)` — idempotency check
- `markWebhookEventProcessed(eventId, type)` — record processed event

**Step 3: Update storage/index.ts**

Keep the existing re-exports of `billing-types` and `billing`. The export surface changes but the barrel file stays the same.

**Step 4: Update supabase-types.ts**

Add TypeScript types for `stripe_customers`, `subscription_items`, and `stripe_webhook_events` tables. Remove types for the 6 dropped tables.

**Step 5: Commit**

```bash
git add -A && git commit -m "feat(billing): rewrite storage layer for stripe — 2 tables replacing 6"
```

---

### Task 4: Stripe webhook handler

**Files:**
- Create: `web/apps/console/src/app/api/stripe/webhook/route.ts`
- Delete: `web/apps/console/src/app/api/razorpay/webhook/route.ts`

**Step 1: Create webhook route**

`web/apps/console/src/app/api/stripe/webhook/route.ts`:
```typescript
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import Stripe from 'stripe'
import { getStripe } from '@/lib/stripe'
import {
  isWebhookEventProcessed,
  markWebhookEventProcessed,
  upsertStripeCustomer,
  updateBillingStatus,
  syncSubscriptionItems,
} from '@/lib/console/storage'

export async function POST(req: NextRequest) {
  const body = await req.text()
  const signature = req.headers.get('stripe-signature')

  if (!signature) {
    return NextResponse.json({ error: 'Missing signature' }, { status: 400 })
  }

  const webhookSecret = process.env.STRIPE_WEBHOOK_SECRET
  if (!webhookSecret) {
    return NextResponse.json({ error: 'Webhook secret not configured' }, { status: 500 })
  }

  let event: Stripe.Event
  try {
    const stripe = getStripe()
    event = stripe.webhooks.constructEvent(body, signature, webhookSecret)
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unknown error'
    return NextResponse.json({ error: `Webhook verification failed: ${message}` }, { status: 400 })
  }

  // Idempotency check
  if (await isWebhookEventProcessed(event.id)) {
    return NextResponse.json({ received: true })
  }

  try {
    switch (event.type) {
      case 'checkout.session.completed':
        await handleCheckoutCompleted(event.data.object as Stripe.Checkout.Session)
        break
      case 'customer.subscription.updated':
        await handleSubscriptionUpdated(event.data.object as Stripe.Subscription)
        break
      case 'customer.subscription.deleted':
        await handleSubscriptionDeleted(event.data.object as Stripe.Subscription)
        break
      case 'invoice.payment_failed':
        await handlePaymentFailed(event.data.object as Stripe.Invoice)
        break
    }

    await markWebhookEventProcessed(event.id, event.type)
  } catch (err) {
    console.error('Webhook handler error:', err)
    return NextResponse.json({ error: 'Handler failed' }, { status: 500 })
  }

  return NextResponse.json({ received: true })
}

async function handleCheckoutCompleted(session: Stripe.Checkout.Session) {
  const installationId = session.metadata?.installation_id
  if (!installationId || !session.customer || !session.subscription) return

  const stripe = getStripe()
  const subscription = await stripe.subscriptions.retrieve(
    session.subscription as string,
    { expand: ['items.data.price.product'] },
  )

  await upsertStripeCustomer({
    installationId,
    stripeCustomerId: session.customer as string,
    stripeSubscriptionId: subscription.id,
    billingStatus: subscription.status === 'active' ? 'active' : 'incomplete',
    currentPeriodEnd: new Date(subscription.current_period_end * 1000).toISOString(),
  })

  await syncItemsFromSubscription(installationId, subscription)
}

async function handleSubscriptionUpdated(subscription: Stripe.Subscription) {
  const customerId = subscription.customer as string
  // Look up installation by stripe_customer_id (need a storage function for this)
  // Update billing_status, current_period_end, payment_issue
  // Sync subscription_items

  const status = subscription.cancel_at_period_end ? 'cancelled' : subscription.status
  const billingStatus = mapStripeStatus(status)

  await updateBillingStatus(
    customerId,
    billingStatus,
    new Date(subscription.current_period_end * 1000).toISOString(),
    false, // clear payment_issue on successful update
  )

  // Fetch full subscription with expanded products for item sync
  const stripe = getStripe()
  const fullSub = await stripe.subscriptions.retrieve(subscription.id, {
    expand: ['items.data.price.product'],
  })

  const stripeCustomer = await getStripeCustomerByCustomerId(customerId)
  if (stripeCustomer) {
    await syncItemsFromSubscription(stripeCustomer.installationId, fullSub)
  }
}

async function handleSubscriptionDeleted(subscription: Stripe.Subscription) {
  const customerId = subscription.customer as string
  await updateBillingStatus(customerId, 'cancelled', null, false)
  // Optionally: delete all subscription_items for this installation
}

async function handlePaymentFailed(invoice: Stripe.Invoice) {
  if (!invoice.customer) return
  await updateBillingStatus(invoice.customer as string, 'past_due', null, true)
}

function mapStripeStatus(status: string): 'active' | 'past_due' | 'cancelled' | 'trialing' | 'incomplete' {
  switch (status) {
    case 'active': return 'active'
    case 'past_due': return 'past_due'
    case 'canceled':
    case 'cancelled': return 'cancelled'
    case 'trialing': return 'trialing'
    default: return 'incomplete'
  }
}

async function syncItemsFromSubscription(installationId: string, subscription: Stripe.Subscription) {
  const items = subscription.items.data.map((item) => {
    const product = item.price.product as Stripe.Product
    const tierMatch = product.metadata?.tier
    return {
      stripeSubscriptionItemId: item.id,
      stripePriceId: item.price.id,
      tier: tierMatch ? parseInt(tierMatch, 10) : 0,
      productName: product.name,
      quantity: item.quantity ?? 1,
    }
  })

  await syncSubscriptionItems(installationId, items)
}
```

Note: `getStripeCustomerByCustomerId` needs to be added to the storage layer (Task 3). Add it during implementation.

**Step 2: Delete Razorpay webhook**

```bash
rm -rf web/apps/console/src/app/api/razorpay/
```

**Step 3: Commit**

```bash
git add -A && git commit -m "feat(billing): add stripe webhook handler, remove razorpay webhook"
```

---

### Task 5: Server actions — checkout & portal sessions

**Files:**
- Create: `web/apps/console/src/app/actions/billing/checkout.ts`
- Rewrite: `web/apps/console/src/app/actions/billing/queries.ts`

**Step 1: Create checkout.ts**

```typescript
'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getStripe } from '@/lib/stripe'
import { getStripeCustomer, upsertStripeCustomer, getMemberRole, getInstallationById } from '@/lib/console/storage'

interface CheckoutAllocation {
  priceId: string
  quantity: number
}

export async function createCheckoutSession(
  installationId: string,
  allocations: CheckoutAllocation[],
) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const installation = await getInstallationById(installationId)
  if (!installation || installation.userId !== session.user.id) {
    throw new Error('Forbidden: only installation owner can manage billing')
  }

  const stripe = getStripe()

  // Get or create Stripe customer
  let stripeCustomer = await getStripeCustomer(installationId)
  let customerId: string

  if (stripeCustomer?.stripeCustomerId) {
    customerId = stripeCustomer.stripeCustomerId
  } else {
    const customer = await stripe.customers.create({
      email: session.user.email,
      name: session.user.name ?? undefined,
      metadata: { installation_id: installationId },
    })
    customerId = customer.id
    await upsertStripeCustomer({
      installationId,
      stripeCustomerId: customerId,
      stripeSubscriptionId: null,
      billingStatus: 'incomplete',
      currentPeriodEnd: null,
    })
  }

  const lineItems = allocations
    .filter((a) => a.quantity > 0)
    .map((a) => ({
      price: a.priceId,
      quantity: a.quantity,
    }))

  const checkoutSession = await stripe.checkout.sessions.create({
    customer: customerId,
    mode: 'subscription',
    line_items: lineItems,
    success_url: `${process.env.NEXT_PUBLIC_APP_URL}/installations/${installationId}?checkout=success`,
    cancel_url: `${process.env.NEXT_PUBLIC_APP_URL}/installations/${installationId}/billing?checkout=cancelled`,
    metadata: { installation_id: installationId },
    subscription_data: {
      metadata: { installation_id: installationId },
    },
  })

  return { url: checkoutSession.url }
}

export async function createPortalSession(installationId: string) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  if (!role && installation?.userId !== session.user.id) {
    throw new Error('Forbidden')
  }

  const stripeCustomer = await getStripeCustomer(installationId)
  if (!stripeCustomer?.stripeCustomerId) {
    throw new Error('No billing account found. Please subscribe first.')
  }

  const stripe = getStripe()
  const portalSession = await stripe.billingPortal.sessions.create({
    customer: stripeCustomer.stripeCustomerId,
    return_url: `${process.env.NEXT_PUBLIC_APP_URL}/installations/${installationId}/billing`,
  })

  return { url: portalSession.url }
}
```

**Step 2: Rewrite queries.ts**

Replace `getRazorpayKey()` with `getStripePublishableKey()`. Replace `fetchSubscriptions` and `fetchInvoices` with functions that read from local `stripe_customers` + `subscription_items` tables:

```typescript
'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import {
  getStripeCustomer,
  getSubscriptionItems,
  getMemberRole,
  getInstallationById,
} from '@/lib/console/storage'
import type { StripeCustomer, SubscriptionItem } from '@/lib/console/storage'

export async function getStripePublishableKey(): Promise<string> {
  const key = process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY
  if (!key) throw new Error('NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY not configured')
  return key
}

export async function fetchBillingStatus(installationId: string): Promise<{
  customer: StripeCustomer | null
  items: SubscriptionItem[]
}> {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const role = await getMemberRole(installationId, session.user.id)
  const installation = await getInstallationById(installationId)
  if (!role && installation?.userId !== session.user.id) {
    throw new Error('Forbidden')
  }

  const [customer, items] = await Promise.all([
    getStripeCustomer(installationId),
    getSubscriptionItems(installationId),
  ])

  return { customer, items }
}
```

**Step 3: Commit**

```bash
git add -A && git commit -m "feat(billing): add stripe checkout and portal session actions"
```

---

### Task 6: Server actions — subscription modifications

**Files:**
- Rewrite: `web/apps/console/src/app/actions/billing/subscriptions.ts`
- Delete: `web/apps/console/src/app/actions/billing/proration.ts`
- Delete: `web/apps/console/src/app/actions/billing/verification.ts`

**Step 1: Rewrite subscriptions.ts**

```typescript
'use server'

import { redirect } from 'next/navigation'
import { getRegistrationSession } from '@/lib/console-auth'
import { getStripe } from '@/lib/stripe'
import { getStripeCustomer, getInstallationById } from '@/lib/console/storage'

interface SubscriptionModification {
  priceId: string
  quantity: number
}

export async function modifySubscription(
  installationId: string,
  modifications: SubscriptionModification[],
) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const installation = await getInstallationById(installationId)
  if (!installation || installation.userId !== session.user.id) {
    throw new Error('Forbidden: only installation owner can modify billing')
  }

  const stripeCustomer = await getStripeCustomer(installationId)
  if (!stripeCustomer?.stripeSubscriptionId) {
    throw new Error('No active subscription found')
  }

  const stripe = getStripe()
  const subscription = await stripe.subscriptions.retrieve(stripeCustomer.stripeSubscriptionId)

  // Build items array: update existing, add new, remove zeroed
  const items: Array<{
    id?: string
    price?: string
    quantity?: number
    deleted?: boolean
  }> = []

  for (const mod of modifications) {
    const existing = subscription.items.data.find((i) => i.price.id === mod.priceId)

    if (existing) {
      if (mod.quantity === 0) {
        items.push({ id: existing.id, deleted: true })
      } else {
        items.push({ id: existing.id, quantity: mod.quantity })
      }
    } else if (mod.quantity > 0) {
      items.push({ price: mod.priceId, quantity: mod.quantity })
    }
  }

  if (items.length === 0) {
    throw new Error('No changes to apply')
  }

  await stripe.subscriptions.update(stripeCustomer.stripeSubscriptionId, {
    items,
    proration_behavior: 'always_invoice',
  })

  return { success: true }
}

export async function cancelSubscription(installationId: string) {
  const session = await getRegistrationSession()
  if (!session?.user) redirect('/login')

  const installation = await getInstallationById(installationId)
  if (!installation || installation.userId !== session.user.id) {
    throw new Error('Forbidden: only installation owner can cancel billing')
  }

  const stripeCustomer = await getStripeCustomer(installationId)
  if (!stripeCustomer?.stripeSubscriptionId) {
    throw new Error('No active subscription found')
  }

  const stripe = getStripe()
  await stripe.subscriptions.update(stripeCustomer.stripeSubscriptionId, {
    cancel_at_period_end: true,
  })

  return { success: true }
}
```

**Step 2: Delete old files**

```bash
rm web/apps/console/src/app/actions/billing/proration.ts
rm web/apps/console/src/app/actions/billing/verification.ts
```

**Step 3: Create barrel re-export**

Verify `web/apps/console/src/app/actions/billing/` has a barrel file. If not, check how existing actions are imported (they may be imported directly by path). Update imports accordingly.

**Step 4: Commit**

```bash
git add -A && git commit -m "feat(billing): add stripe subscription modification, delete proration and verification"
```

---

### Task 7: Hook — replace payment interactions

**Files:**
- Rewrite: `web/apps/console/src/hooks/use-subscription-payments.ts`

**Step 1: Rewrite the hook**

Replace the 204-line Razorpay hook with a ~60-line Stripe version:

```typescript
'use client'

import { useCallback, useState } from 'react'
import { useRouter } from 'next/navigation'
import { toast } from 'sonner'
import { createCheckoutSession, createPortalSession } from '@/app/actions/billing/checkout'
import { modifySubscription } from '@/app/actions/billing/subscriptions'

interface UseSubscriptionPaymentsOptions {
  installationId: string
}

export function useSubscriptionPayments({ installationId }: UseSubscriptionPaymentsOptions) {
  const router = useRouter()
  const [loading, setLoading] = useState(false)

  const handleSubscribe = useCallback(
    async (allocations: { priceId: string; quantity: number }[]) => {
      setLoading(true)
      try {
        const { url } = await createCheckoutSession(installationId, allocations)
        if (url) window.location.href = url
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to start checkout')
        setLoading(false)
      }
    },
    [installationId],
  )

  const handleModify = useCallback(
    async (modifications: { priceId: string; quantity: number }[]) => {
      setLoading(true)
      try {
        await modifySubscription(installationId, modifications)
        toast.success('Subscription updated. Prorated charges applied.')
        router.refresh()
      } catch (error) {
        toast.error(error instanceof Error ? error.message : 'Failed to modify subscription')
      } finally {
        setLoading(false)
      }
    },
    [installationId, router],
  )

  const handleManageBilling = useCallback(async () => {
    setLoading(true)
    try {
      const { url } = await createPortalSession(installationId)
      if (url) window.location.href = url
    } catch (error) {
      toast.error(error instanceof Error ? error.message : 'Failed to open billing portal')
      setLoading(false)
    }
  }, [installationId])

  return {
    loading,
    handleSubscribe,
    handleModify,
    handleManageBilling,
  }
}
```

**Step 2: Commit**

```bash
git add -A && git commit -m "feat(billing): rewrite payment hook for stripe — 60 lines replacing 204"
```

---

### Task 8: Simplified billing UI

**Files:**
- Simplify: `web/apps/console/src/components/billing/subscription-management.tsx`
- Simplify: `web/apps/console/src/components/billing/subscription-configurator.tsx`
- Simplify: `web/apps/console/src/components/billing/subscription-header.tsx`
- Simplify: `web/apps/console/src/components/billing/subscription-status.tsx`
- Create: `web/apps/console/src/components/billing/payment-warning-banner.tsx`
- Delete: `web/apps/console/src/components/billing/payment-due-banner.tsx`
- Delete: `web/apps/console/src/components/billing/invoice-history.tsx`
- Delete: `web/apps/console/src/components/billing/past-subscriptions.tsx`

**Step 1: Create payment-warning-banner.tsx**

Simple banner shown when `payment_issue = true`:
```typescript
'use client'

import { AlertTriangle } from 'lucide-react'
import { Button } from '@kloudlite/ui/button'

interface PaymentWarningBannerProps {
  onManageBilling: () => void
}

export function PaymentWarningBanner({ onManageBilling }: PaymentWarningBannerProps) {
  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-800 dark:bg-amber-950">
      <div className="flex items-center gap-3">
        <AlertTriangle className="h-5 w-5 text-amber-600 dark:text-amber-400" />
        <div className="flex-1">
          <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
            Payment issue detected
          </p>
          <p className="text-sm text-amber-700 dark:text-amber-300">
            Please update your payment method to avoid service interruption.
          </p>
        </div>
        <Button variant="outline" size="sm" onClick={onManageBilling}>
          Update Payment Method
        </Button>
      </div>
    </div>
  )
}
```

**Step 2: Simplify subscription-management.tsx**

Rewrite to use new types (`StripeCustomer`, `SubscriptionItem`) instead of `Plan`, `Subscription`, `Invoice`. Remove `PaymentDueBanner`, `PastSubscriptions`, `InvoiceHistory`. Add "Manage Billing" button. Use the simplified hook.

**Step 3: Simplify subscription-configurator.tsx**

Remove proration preview logic. Remove billing period toggle (monthly only for now). The configurator now works with Stripe price IDs instead of plan IDs. Show tier name + seat stepper + price per seat. "Save Changes" calls `handleModify` with Stripe price IDs and quantities.

**Step 4: Simplify subscription-status.tsx and subscription-header.tsx**

Update to use `SubscriptionItem[]` instead of `Subscription` + `Plan`. Show product name, quantity, per-unit cost (read from a config constant or Stripe metadata), next billing date.

**Step 5: Delete old components**

```bash
rm web/apps/console/src/components/billing/payment-due-banner.tsx
rm web/apps/console/src/components/billing/invoice-history.tsx
rm web/apps/console/src/components/billing/past-subscriptions.tsx
```

**Step 6: Commit**

```bash
git add -A && git commit -m "feat(billing): simplify billing UI — delete 3 components, add payment warning banner"
```

---

### Task 9: Installation creation flow

**Files:**
- Modify: `web/apps/console/src/components/kl-cloud-installation-form.tsx`
- Delete: `web/apps/console/src/components/razorpay-provider.tsx`

**Step 1: Update kl-cloud-installation-form.tsx**

Replace the Razorpay checkout step with a Stripe Checkout redirect. Instead of opening a popup after installation creation, redirect to Stripe Checkout. The webhook will activate the installation on successful payment.

Key changes:
- Remove `useRazorpay` import and `RazorpayProvider` wrapper
- Add a plan selection step using the simplified configurator
- On submit: create installation → create Stripe Checkout Session → redirect

**Step 2: Delete razorpay-provider.tsx**

```bash
rm web/apps/console/src/components/razorpay-provider.tsx
```

**Step 3: Commit**

```bash
git add -A && git commit -m "feat(billing): update installation form for stripe checkout redirect"
```

---

### Task 10: Billing page & CSP headers

**Files:**
- Modify: `web/apps/console/src/app/installations/[id]/billing/page.tsx`
- Modify: `web/apps/console/src/proxy.ts`

**Step 1: Update billing page**

Remove `RazorpayProvider` wrapper, `InvoiceHistory` section. Fetch `stripe_customers` + `subscription_items` instead of plans/subscriptions/invoices. Pass new props to `SubscriptionManagement`.

**Step 2: Update CSP headers in proxy.ts**

Replace Razorpay domains with Stripe:
```typescript
const scriptSrc = [
  "'self'",
  "'unsafe-inline'",
  ...(process.env.NODE_ENV === 'development' ? ["'unsafe-eval'"] : []),
  'https://challenges.cloudflare.com',
  'https://static.cloudflareinsights.com',
  'https://js.stripe.com',
].join(' ')

// connect-src: add https://api.stripe.com
// frame-src: add https://js.stripe.com
```

**Step 3: Commit**

```bash
git add -A && git commit -m "fix(billing): update billing page and CSP headers for stripe"
```

---

### Task 11: Delete billing cron & CI cleanup

**Files:**
- Delete: `supabase/functions/billing-cron/index.ts` (and entire directory)
- Modify: `.github/workflows/deploy-edge-functions.yml`

**Step 1: Delete billing cron**

```bash
rm -rf supabase/functions/billing-cron/
```

**Step 2: Update CI workflow**

Remove `RAZORPAY_KEY_ID` and `RAZORPAY_KEY_SECRET` secrets from the edge function deployment workflow. If `billing-cron` is the only edge function, consider whether the workflow is still needed.

**Step 3: Commit**

```bash
git add -A && git commit -m "chore(billing): remove billing cron and razorpay CI secrets"
```

---

### Task 12: Update tests & billing utils

**Files:**
- Delete: `web/apps/console/src/app/actions/billing/proration.test.ts`
- Modify: `web/apps/console/src/lib/billing-utils.ts`
- Modify: `web/apps/console/src/lib/billing-utils.test.ts`

**Step 1: Delete proration tests**

```bash
rm web/apps/console/src/app/actions/billing/proration.test.ts
```

**Step 2: Update billing-utils.ts**

Change default currency from `'INR'` to `'USD'`:
```typescript
export function formatCurrency(amountInCents: number, currency: string = 'USD'): string {
  return `${getCurrencySymbol(currency)}${(amountInCents / 100).toFixed(2)}`
}
```

**Step 3: Update billing-utils.test.ts**

Update test expectations for USD default.

**Step 4: Run tests**

```bash
cd web/apps/console && bunx vitest run
```

**Step 5: Commit**

```bash
git add -A && git commit -m "test(billing): update billing utils for USD default, delete proration tests"
```

---

### Task 13: Build & lint verification

**Step 1: Run build**

```bash
cd web && bun run build:console
```

Fix any type errors or import issues from the migration.

**Step 2: Run lint**

```bash
cd web && bun run lint
```

Fix any lint errors.

**Step 3: Run tests**

```bash
cd web/apps/console && bunx vitest run
```

**Step 4: Final commit if fixes needed**

```bash
git add -A && git commit -m "fix(billing): resolve build and lint issues from stripe migration"
```
