# Pay-As-You-Go Prepaid Credits Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace subscription-based billing with pay-as-you-go prepaid credits where users top up a credit balance and are charged hourly for active resources (Control Plane, WorkMachines, Storage).

**Architecture:** Our own credit ledger in Supabase/PostgreSQL. Stripe used only for payment collection (invoices in `payment` mode for top-ups) and informational monthly summaries. Usage events flow from the in-cluster WorkMachine controller → Console API → usage tables. A cron job debits credits every 5 minutes and enforces balance limits. Heartbeat reconciliation via verify-key ensures eventual consistency.

**Tech Stack:** Next.js 16, React 19, TypeScript 5, Supabase/PostgreSQL, Stripe SDK (invoices only), Go 1.24 (WorkMachine controller), Bun, Tailwind 4, shadcn/ui, Vitest, Playwright.

**Design doc:** `docs/plans/2026-03-09-pay-as-you-go-billing-design.md`

---

## Task 1: Database Migration — Credit & Usage Tables

**Files:**
- Create: `web/apps/console/src/lib/console/migrations/007_credits_billing.sql`

**Step 1: Write the migration SQL**

Create file `web/apps/console/src/lib/console/migrations/007_credits_billing.sql`:

```sql
-- Pay-as-you-go credits billing tables
-- Replaces subscription-based billing with prepaid credit system

-- ============================================
-- credit_accounts: one per org, tracks balance
-- ============================================
CREATE TABLE IF NOT EXISTS credit_accounts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL UNIQUE REFERENCES organizations(id) ON DELETE CASCADE,
  balance NUMERIC(12, 4) NOT NULL DEFAULT 0,
  auto_topup_enabled BOOLEAN NOT NULL DEFAULT false,
  auto_topup_threshold NUMERIC(10, 2),
  auto_topup_amount NUMERIC(10, 2),
  stripe_customer_id TEXT UNIQUE,
  negative_balance_flagged BOOLEAN NOT NULL DEFAULT false,
  low_balance_warning BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ============================================
-- credit_transactions: append-only ledger
-- ============================================
CREATE TABLE IF NOT EXISTS credit_transactions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  type TEXT NOT NULL CHECK (type IN ('topup', 'usage_debit', 'adjustment')),
  amount NUMERIC(12, 4) NOT NULL,
  description TEXT,
  stripe_invoice_id TEXT,
  usage_period_id UUID,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_credit_transactions_org_id ON credit_transactions(org_id);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_created_at ON credit_transactions(created_at);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_type ON credit_transactions(type);

-- ============================================
-- usage_events: raw events from controllers
-- ============================================
CREATE TABLE IF NOT EXISTS usage_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  event_type TEXT NOT NULL CHECK (event_type IN (
    'workmachine.started', 'workmachine.stopped', 'workmachine.resized',
    'controlplane.started', 'controlplane.stopped',
    'storage.provisioned', 'storage.resized', 'storage.deleted'
  )),
  resource_id TEXT NOT NULL,
  resource_type TEXT,
  metadata JSONB DEFAULT '{}',
  event_timestamp TIMESTAMPTZ NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(installation_id, resource_id, event_type, event_timestamp)
);

CREATE INDEX IF NOT EXISTS idx_usage_events_installation_id ON usage_events(installation_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_event_type ON usage_events(event_type);

-- ============================================
-- usage_periods: tracks active resource usage
-- ============================================
CREATE TABLE IF NOT EXISTS usage_periods (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  installation_id UUID NOT NULL REFERENCES installations(id) ON DELETE CASCADE,
  org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
  resource_id TEXT NOT NULL,
  resource_type TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL,
  ended_at TIMESTAMPTZ,
  hourly_rate NUMERIC(10, 6) NOT NULL,
  total_cost NUMERIC(12, 4) NOT NULL DEFAULT 0,
  last_billed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_usage_periods_active ON usage_periods(org_id) WHERE ended_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_usage_periods_installation_id ON usage_periods(installation_id);
CREATE INDEX IF NOT EXISTS idx_usage_periods_org_id ON usage_periods(org_id);

-- ============================================
-- pricing_tiers: configurable resource pricing
-- ============================================
CREATE TABLE IF NOT EXISTS pricing_tiers (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  resource_type TEXT NOT NULL UNIQUE,
  display_name TEXT NOT NULL,
  hourly_rate NUMERIC(10, 6) NOT NULL,
  unit TEXT NOT NULL DEFAULT 'hour',
  category TEXT NOT NULL CHECK (category IN ('compute', 'storage')),
  specs JSONB DEFAULT '{}',
  is_active BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed pricing tiers
INSERT INTO pricing_tiers (resource_type, display_name, hourly_rate, unit, category, specs) VALUES
  ('controlplane', 'Control Plane', 0.020000, 'hour', 'compute', '{"description": "Master node, always running"}'),
  ('workmachine.standard', 'Standard WorkMachine', 0.050000, 'hour', 'compute', '{"vcpu": 2, "memory_gb": 4}'),
  ('workmachine.performance', 'Performance WorkMachine', 0.120000, 'hour', 'compute', '{"vcpu": 4, "memory_gb": 8}'),
  ('storage.vm', 'VM Volume Storage', 0.000056, 'gb_hour', 'storage', '{"monthly_rate_per_gb": 0.04}'),
  ('storage.object', 'Object Storage', 0.000056, 'gb_hour', 'storage', '{"monthly_rate_per_gb": 0.04}')
ON CONFLICT (resource_type) DO NOTHING;

-- ============================================
-- Triggers for updated_at
-- ============================================
CREATE TRIGGER update_credit_accounts_updated_at
  BEFORE UPDATE ON credit_accounts
  FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- RLS Policies
-- ============================================
ALTER TABLE credit_accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE credit_transactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_periods ENABLE ROW LEVEL SECURITY;
ALTER TABLE pricing_tiers ENABLE ROW LEVEL SECURITY;

-- Service role (our API) gets full access
CREATE POLICY "Service role full access" ON credit_accounts FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Service role full access" ON credit_transactions FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Service role full access" ON usage_events FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Service role full access" ON usage_periods FOR ALL USING (true) WITH CHECK (true);
CREATE POLICY "Service role full access" ON pricing_tiers FOR ALL USING (true) WITH CHECK (true);

-- ============================================
-- RPC: Atomic credit debit
-- ============================================
CREATE OR REPLACE FUNCTION debit_credits(
  p_org_id UUID,
  p_amount NUMERIC,
  p_description TEXT,
  p_usage_period_id UUID DEFAULT NULL
)
RETURNS NUMERIC
LANGUAGE plpgsql
SET search_path = ''
AS $$
DECLARE
  v_new_balance NUMERIC;
BEGIN
  UPDATE public.credit_accounts
  SET balance = balance - p_amount
  WHERE org_id = p_org_id
  RETURNING balance INTO v_new_balance;

  INSERT INTO public.credit_transactions (org_id, type, amount, description, usage_period_id)
  VALUES (p_org_id, 'usage_debit', -p_amount, p_description, p_usage_period_id);

  RETURN v_new_balance;
END;
$$;

-- ============================================
-- RPC: Atomic credit top-up
-- ============================================
CREATE OR REPLACE FUNCTION topup_credits(
  p_org_id UUID,
  p_amount NUMERIC,
  p_description TEXT,
  p_stripe_invoice_id TEXT DEFAULT NULL
)
RETURNS NUMERIC
LANGUAGE plpgsql
SET search_path = ''
AS $$
DECLARE
  v_new_balance NUMERIC;
BEGIN
  UPDATE public.credit_accounts
  SET balance = balance + p_amount,
      low_balance_warning = false,
      negative_balance_flagged = false
  WHERE org_id = p_org_id
  RETURNING balance INTO v_new_balance;

  INSERT INTO public.credit_transactions (org_id, type, amount, description, stripe_invoice_id)
  VALUES (p_org_id, 'topup', p_amount, p_description, p_stripe_invoice_id);

  RETURN v_new_balance;
END;
$$;
```

**Step 2: Run the migration against local Supabase**

```bash
cd web/apps/console
# The migration runner is invoked on app startup or manually
# Apply via Supabase SQL editor or psql
```

Verify tables exist with a quick query: `SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('credit_accounts', 'credit_transactions', 'usage_events', 'usage_periods', 'pricing_tiers');`

**Step 3: Commit**

```bash
git add web/apps/console/src/lib/console/migrations/007_credits_billing.sql
git commit -m "feat(billing): add credit accounts, transactions, usage events, and pricing tables"
```

---

## Task 2: Storage Layer — Credit & Usage Types

**Files:**
- Create: `web/apps/console/src/lib/console/storage/credits-types.ts`

**Step 1: Write the types**

```typescript
// Credit account for an org
export interface CreditAccount {
  id: string;
  orgId: string;
  balance: number;
  autoTopupEnabled: boolean;
  autoTopupThreshold: number | null;
  autoTopupAmount: number | null;
  stripeCustomerId: string | null;
  negativeBalanceFlagged: boolean;
  lowBalanceWarning: boolean;
  createdAt: string;
  updatedAt: string;
}

// Append-only credit ledger entry
export interface CreditTransaction {
  id: string;
  orgId: string;
  type: 'topup' | 'usage_debit' | 'adjustment';
  amount: number;
  description: string | null;
  stripeInvoiceId: string | null;
  usagePeriodId: string | null;
  createdAt: string;
}

// Raw usage event from controller
export interface UsageEvent {
  id: string;
  installationId: string;
  eventType: string;
  resourceId: string;
  resourceType: string | null;
  metadata: Record<string, unknown>;
  eventTimestamp: string;
  createdAt: string;
}

// Active or closed usage period
export interface UsagePeriod {
  id: string;
  installationId: string;
  orgId: string;
  resourceId: string;
  resourceType: string;
  startedAt: string;
  endedAt: string | null;
  hourlyRate: number;
  totalCost: number;
  lastBilledAt: string;
  createdAt: string;
}

// Pricing tier configuration
export interface PricingTier {
  id: string;
  resourceType: string;
  displayName: string;
  hourlyRate: number;
  unit: string;
  category: 'compute' | 'storage';
  specs: Record<string, unknown>;
  isActive: boolean;
  createdAt: string;
}
```

**Step 2: Commit**

```bash
git add web/apps/console/src/lib/console/storage/credits-types.ts
git commit -m "feat(billing): add credit and usage type definitions"
```

---

## Task 3: Storage Layer — Credits CRUD Functions

**Files:**
- Create: `web/apps/console/src/lib/console/storage/credits.ts`

**Step 1: Write the storage functions**

Model after `web/apps/console/src/lib/console/storage/billing.ts` (line 1-270) for patterns: use `createClient()` from `@/lib/supabase/server`, map snake_case rows to camelCase types.

Functions to implement:

```typescript
import { createClient } from '@/lib/supabase/server';
import type { CreditAccount, CreditTransaction, UsageEvent, UsagePeriod, PricingTier } from './credits-types';

// --- Mapping helpers ---
function mapToCreditAccount(row: any): CreditAccount { /* snake_case to camelCase */ }
function mapToCreditTransaction(row: any): CreditTransaction { /* ... */ }
function mapToUsageEvent(row: any): UsageEvent { /* ... */ }
function mapToUsagePeriod(row: any): UsagePeriod { /* ... */ }
function mapToPricingTier(row: any): PricingTier { /* ... */ }

// --- Credit Accounts ---
export async function getCreditAccount(orgId: string): Promise<CreditAccount | null>
export async function getCreditAccountByCustomerId(stripeCustomerId: string): Promise<CreditAccount | null>
export async function ensureCreditAccount(orgId: string, stripeCustomerId?: string): Promise<CreditAccount>
  // upsert on org_id conflict
export async function updateCreditAccountAutoTopup(orgId: string, enabled: boolean, threshold?: number, amount?: number): Promise<void>
export async function updateCreditAccountWarnings(orgId: string, lowBalanceWarning: boolean, negativeBalanceFlagged?: boolean): Promise<void>
export async function updateCreditAccountStripeCustomer(orgId: string, stripeCustomerId: string): Promise<void>

// --- Credit Transactions ---
export async function getCreditTransactions(orgId: string, limit?: number, offset?: number): Promise<CreditTransaction[]>
export async function getCreditTransactionsByPeriod(orgId: string, startDate: string, endDate: string): Promise<CreditTransaction[]>
export async function topupCredits(orgId: string, amount: number, description: string, stripeInvoiceId?: string): Promise<number>
  // calls RPC topup_credits, returns new balance
export async function debitCredits(orgId: string, amount: number, description: string, usagePeriodId?: string): Promise<number>
  // calls RPC debit_credits, returns new balance

// --- Usage Events ---
export async function insertUsageEvent(event: Omit<UsageEvent, 'id' | 'createdAt'>): Promise<UsageEvent>
  // handles unique constraint violations gracefully (duplicate = no-op)

// --- Usage Periods ---
export async function getActiveUsagePeriods(orgId: string): Promise<UsagePeriod[]>
  // WHERE ended_at IS NULL
export async function getActiveUsagePeriodsForInstallation(installationId: string): Promise<UsagePeriod[]>
export async function openUsagePeriod(params: { installationId: string; orgId: string; resourceId: string; resourceType: string; hourlyRate: number }): Promise<UsagePeriod>
export async function closeUsagePeriod(resourceId: string, installationId: string): Promise<UsagePeriod | null>
  // sets ended_at = now(), calculates total_cost from hours elapsed * hourly_rate
export async function closeAllUsagePeriodsForInstallation(installationId: string): Promise<void>
export async function updateLastBilledAt(periodIds: string[]): Promise<void>
export async function getUsagePeriodsByDateRange(orgId: string, startDate: string, endDate: string): Promise<UsagePeriod[]>

// --- Pricing Tiers ---
export async function getPricingTiers(): Promise<PricingTier[]>
  // WHERE is_active = true, cached in-memory
export async function getPricingTierByType(resourceType: string): Promise<PricingTier | null>
export async function getHourlyRate(resourceType: string): Promise<number>
  // convenience wrapper, throws if not found
```

**Step 2: Commit**

```bash
git add web/apps/console/src/lib/console/storage/credits.ts
git commit -m "feat(billing): add credits storage layer with CRUD functions"
```

---

## Task 4: Usage Event API Route

**Files:**
- Create: `web/apps/console/src/app/api/installations/usage-event/route.ts`

**Step 1: Write the API route**

```typescript
// POST /api/installations/usage-event
// Auth: x-installation-key header
// Body: { event_type, resource_id, resource_type, metadata, timestamp }
```

Logic:
1. Validate `x-installation-key` header — look up installation via `getInstallationByKey()`
2. Get org_id from installation
3. Validate event_type is one of the allowed types
4. Insert into `usage_events`
5. For `started`/`provisioned` events:
   - Look up hourly rate from `pricing_tiers` by resource_type
   - Call `openUsagePeriod()` — creates new active period
6. For `stopped`/`deleted` events:
   - Call `closeUsagePeriod()` — closes active period, calculates cost
   - Call `debitCredits()` — debit the calculated cost
7. For `resized` events:
   - Close current period (debit accrued cost)
   - Open new period with new resource_type/hourly_rate
8. Return `{ success: true }`

Handle errors: duplicate events (unique constraint) → return 200 OK (idempotent). Invalid key → 401. Invalid event_type → 400.

**Step 2: Commit**

```bash
git add web/apps/console/src/app/api/installations/usage-event/route.ts
git commit -m "feat(billing): add usage event ingestion API route"
```

---

## Task 5: Extend verify-key for Heartbeat Reconciliation

**Files:**
- Modify: `web/apps/console/src/app/api/installations/verify-key/route.ts` (85 lines)

**Step 1: Extend the request body**

Add optional fields to the request body parsing:
```typescript
const { installationKey, provider, region, running_machines, volumes } = await req.json();
```

**Step 2: Add reconciliation logic after existing health check update**

After `updateHealthCheck()` (around line 70), add:

```typescript
if (running_machines || volumes) {
  await reconcileUsagePeriods(installation.id, installation.orgId, running_machines ?? [], volumes ?? []);
}
```

**Step 3: Write reconcileUsagePeriods function**

Create helper (in same file or in a shared util):

```typescript
async function reconcileUsagePeriods(
  installationId: string,
  orgId: string,
  runningMachines: Array<{ machine_id: string; machine_type: string; started_at: string }>,
  volumes: Array<{ volume_id: string; volume_type: 'vm' | 'object'; size_gb: number; created_at: string }>
) {
  const activePeriods = await getActiveUsagePeriodsForInstallation(installationId);

  // Build sets for quick lookup
  const reportedResourceIds = new Set([
    ...runningMachines.map(m => m.machine_id),
    ...volumes.map(v => v.volume_id),
  ]);
  const activeResourceIds = new Set(activePeriods.map(p => p.resourceId));

  // 1. Resources running but no open period → create one (missed start event)
  for (const machine of runningMachines) {
    if (!activeResourceIds.has(machine.machine_id)) {
      const rate = await getHourlyRate(machine.machine_type);
      await openUsagePeriod({
        installationId, orgId,
        resourceId: machine.machine_id,
        resourceType: machine.machine_type,
        hourlyRate: rate,
      });
    }
  }
  for (const volume of volumes) {
    if (!activeResourceIds.has(volume.volume_id)) {
      const storageType = volume.volume_type === 'vm' ? 'storage.vm' : 'storage.object';
      const rate = await getHourlyRate(storageType);
      // For storage, rate is per GB-hour, so multiply by size
      await openUsagePeriod({
        installationId, orgId,
        resourceId: volume.volume_id,
        resourceType: storageType,
        hourlyRate: rate * volume.size_gb,
      });
    }
  }

  // 2. Open period but resource not reported → close it (missed stop event)
  // Skip control plane resources (they don't appear in heartbeat)
  for (const period of activePeriods) {
    if (!reportedResourceIds.has(period.resourceId) && !period.resourceType.startsWith('controlplane')) {
      await closeUsagePeriod(period.resourceId, installationId);
      // Debit the cost
      if (period.totalCost > 0) {
        await debitCredits(orgId, period.totalCost, `Usage: ${period.resourceType} ${period.resourceId}`);
      }
    }
  }
}
```

**Step 4: Commit**

```bash
git add web/apps/console/src/app/api/installations/verify-key/route.ts
git commit -m "feat(billing): extend verify-key with heartbeat usage reconciliation"
```

---

## Task 6: Credit Top-Up API Route (Stripe Invoice)

**Files:**
- Create: `web/apps/console/src/app/api/orgs/[orgId]/billing/topup/route.ts`

**Step 1: Write the top-up route**

```typescript
// POST /api/orgs/[orgId]/billing/topup
// Body: { amount: number }
// Auth: org owner only
```

Logic:
1. Validate session, org ownership
2. Validate amount >= calculated minimum (or a floor like $5)
3. Get or create `credit_account` for this org
4. Get or create Stripe customer
5. Create Stripe Invoice:
   ```typescript
   const invoice = await stripe.invoices.create({
     customer: stripeCustomerId,
     collection_method: 'send_invoice',
     days_until_due: 0,
     auto_advance: true,
     metadata: { org_id: orgId, type: 'credit_topup' },
   });
   await stripe.invoiceItems.create({
     customer: stripeCustomerId,
     invoice: invoice.id,
     amount: Math.round(amount * 100), // cents
     currency: 'usd',
     description: `Kloudlite Credit Top-Up: $${amount.toFixed(2)}`,
   });
   const finalizedInvoice = await stripe.invoices.finalizeInvoice(invoice.id);
   ```
6. Return `{ url: finalizedInvoice.hosted_invoice_url }` → redirect user to Stripe-hosted payment page

**Step 2: Commit**

```bash
git add web/apps/console/src/app/api/orgs/[orgId]/billing/topup/route.ts
git commit -m "feat(billing): add credit top-up API route using Stripe invoices"
```

---

## Task 7: Credit Balance & Transactions API Route

**Files:**
- Create: `web/apps/console/src/app/api/orgs/[orgId]/billing/credits/route.ts`

**Step 1: Write the credits route**

```typescript
// GET /api/orgs/[orgId]/billing/credits
// Returns: { account: CreditAccount, transactions: CreditTransaction[], activePeriods: UsagePeriod[] }
// Auth: org member (any role)
```

Logic:
1. Validate session, org membership
2. Fetch credit account, recent transactions (last 50), active usage periods
3. Return combined data

Also add:

```typescript
// PATCH /api/orgs/[orgId]/billing/credits
// Body: { autoTopupEnabled, autoTopupThreshold, autoTopupAmount }
// Auth: org owner only
```

For updating auto top-up settings.

**Step 2: Commit**

```bash
git add web/apps/console/src/app/api/orgs/[orgId]/billing/credits/route.ts
git commit -m "feat(billing): add credits balance and auto-topup settings API route"
```

---

## Task 8: Pricing Tiers API Route

**Files:**
- Create: `web/apps/console/src/app/api/pricing/route.ts`

**Step 1: Write the pricing route**

```typescript
// GET /api/pricing
// Returns: { tiers: PricingTier[] }
// Auth: none (public)
```

Returns all active pricing tiers. Used by the installation form to display hourly rates and calculate projected costs.

**Step 2: Commit**

```bash
git add web/apps/console/src/app/api/pricing/route.ts
git commit -m "feat(billing): add public pricing tiers API route"
```

---

## Task 9: Balance Checker Cron Job

**Files:**
- Create: `web/apps/console/src/app/api/cron/balance-checker/route.ts`

**Step 1: Write the cron route**

```typescript
// POST /api/cron/balance-checker
// Auth: cron secret header (x-cron-secret)
// Called every 5 minutes by external cron (Supabase pg_cron, Vercel Cron, or simple setInterval)
```

Logic:
1. Query all orgs with active usage periods (`SELECT DISTINCT org_id FROM usage_periods WHERE ended_at IS NULL`)
2. For each org:
   a. Get all active usage periods
   b. Calculate cost accrued since `last_billed_at` for each period:
      - Compute hours: `(now - last_billed_at) / 3600`
      - For compute: `hours * hourly_rate`
      - Sum total across all periods
   c. Atomic debit via `debit_credits(org_id, total, 'Periodic usage debit')`
   d. Update `last_billed_at` on all debited periods
   e. Check new balance:
      - Calculate burn rate (cost per hour from all active periods)
      - If projected $0 within 24 hours → set `low_balance_warning = true`
      - If auto_topup_enabled and balance < threshold → trigger auto top-up (create Stripe Invoice)
      - If balance <= 0 → pause WorkMachines: call `POST /api/installations/usage-event` with `workmachine.stopped` for each active WM, or directly close periods and update installation status
      - If balance < -10 → flag org for review, pause control plane too

**Step 2: Commit**

```bash
git add web/apps/console/src/app/api/cron/balance-checker/route.ts
git commit -m "feat(billing): add balance checker cron job for periodic usage debiting"
```

---

## Task 10: Monthly Invoice Cron Job

**Files:**
- Create: `web/apps/console/src/app/api/cron/monthly-invoice/route.ts`

**Step 1: Write the monthly invoice route**

```typescript
// POST /api/cron/monthly-invoice
// Auth: cron secret header
// Called on 1st of each month
```

Logic:
1. Get all orgs with credit accounts
2. For each org, query `credit_transactions` of type `usage_debit` for the previous month
3. Group by resource_type from associated `usage_periods`
4. Create Stripe Invoice (informational, paid = $0):
   ```typescript
   const invoice = await stripe.invoices.create({
     customer: stripeCustomerId,
     collection_method: 'send_invoice',
     auto_advance: false, // don't try to collect payment
     metadata: { org_id: orgId, type: 'monthly_summary', period: '2026-02' },
   });
   // Add line items per resource type (amount = 0, description has breakdown)
   for (const group of usageGroups) {
     await stripe.invoiceItems.create({
       customer: stripeCustomerId,
       invoice: invoice.id,
       amount: 0,
       currency: 'usd',
       description: `${group.resourceType}: ${group.hours.toFixed(1)} hours × $${group.rate}/hr = $${group.total.toFixed(2)}`,
     });
   }
   await stripe.invoices.finalizeInvoice(invoice.id);
   ```

**Step 2: Commit**

```bash
git add web/apps/console/src/app/api/cron/monthly-invoice/route.ts
git commit -m "feat(billing): add monthly informational invoice generation cron job"
```

---

## Task 11: Update Stripe Webhook Handler

**Files:**
- Modify: `web/apps/console/src/app/api/stripe/webhook/route.ts` (233 lines)

**Step 1: Add invoice.paid handler**

Add a new event handler for `invoice.paid`:

```typescript
case 'invoice.paid': {
  const invoice = event.data.object;
  if (invoice.metadata?.type === 'credit_topup') {
    const orgId = invoice.metadata.org_id;
    const amount = invoice.amount_paid / 100; // cents to dollars
    await topupCredits(orgId, amount, `Top-up via Stripe Invoice ${invoice.id}`, invoice.id);
  }
  break;
}
```

**Step 2: Remove/deprecate subscription handlers**

The existing `checkout.session.completed`, `customer.subscription.updated`, and `customer.subscription.deleted` handlers are no longer needed. Keep them temporarily for any active subscriptions during migration, but add a deprecation comment. They can be fully removed once all subscriptions are cancelled.

**Step 3: Commit**

```bash
git add web/apps/console/src/app/api/stripe/webhook/route.ts
git commit -m "feat(billing): add invoice.paid webhook handler for credit top-ups"
```

---

## Task 12: Replace Stripe Bootstrap with Hourly Pricing

**Files:**
- Modify: `web/apps/console/src/lib/stripe-bootstrap.ts` (227 lines)

**Step 1: Simplify the bootstrap**

The old bootstrap creates subscription products with monthly recurring prices. We no longer need Stripe products/prices for billing since we're using custom invoices. The `pricing_tiers` DB table is our source of truth.

Options:
- **Minimal approach:** Remove the subscription product creation entirely. Keep only a helper to create Stripe customers. The pricing info comes from the `pricing_tiers` table.
- Keep the file but gut it to just export `getStripePricing()` that reads from `pricing_tiers` DB table instead of Stripe products.

Replace contents with:

```typescript
import { getPricingTiers } from '@/lib/console/storage/credits';
import type { PricingTier } from '@/lib/console/storage/credits-types';

let cachedTiers: PricingTier[] | null = null;

export async function getActivePricingTiers(): Promise<PricingTier[]> {
  if (cachedTiers) return cachedTiers;
  cachedTiers = await getPricingTiers();
  return cachedTiers;
}

export function clearPricingCache() {
  cachedTiers = null;
}

export function calculateProjectedMonthlyCost(tiers: PricingTier[], selectedResources: Array<{ resourceType: string; quantity?: number; sizeGb?: number }>): number {
  let total = 0;
  for (const resource of selectedResources) {
    const tier = tiers.find(t => t.resourceType === resource.resourceType);
    if (!tier) continue;
    if (tier.category === 'storage') {
      total += tier.hourlyRate * (resource.sizeGb ?? 0) * 24 * 30;
    } else {
      total += tier.hourlyRate * (resource.quantity ?? 1) * 24 * 30;
    }
  }
  return total;
}
```

**Step 2: Commit**

```bash
git add web/apps/console/src/lib/stripe-bootstrap.ts
git commit -m "refactor(billing): replace subscription pricing with hourly pricing from DB"
```

---

## Task 13: Rewrite Installation Form for Pay-As-You-Go

**Files:**
- Modify: `web/apps/console/src/components/kl-cloud-installation-form.tsx` (514 lines)

**Step 1: Remove subscription tier selection**

Remove the tier quantity steppers and subscription checkout flow. Replace with:

1. **Region selector** (keep existing)
2. **WorkMachine configuration** — radio group or cards showing available compute tiers from `pricing_tiers` with hourly rates displayed
3. **Storage size** — slider or input for initial volume size (GB)
4. **Cost summary panel** — calculated projected monthly cost:
   ```
   Control Plane: $0.02/hr (~$14.40/mo)
   1× Standard WorkMachine: $0.05/hr (~$36.00/mo)
   Storage (50 GB): ~$2.00/mo
   ─────────────────────────────
   Estimated total: ~$52.40/mo
   ```
5. **Balance gate:**
   - Fetch credit balance via `GET /api/orgs/${orgId}/billing/credits`
   - If balance >= 30-day projected cost → show "Create Installation" button
   - If balance < 30-day projected cost → show shortfall and "Add Credits ($X.XX minimum)" button
   - "Add Credits" calls `POST /api/orgs/${orgId}/billing/topup` with calculated minimum → redirects to Stripe invoice payment page
   - After payment, redirect back with `?continue=installationId` to resume creation

**Step 2: Update the submit handler**

Instead of creating a Stripe Checkout session, the submit now:
1. Creates the installation (existing `POST /api/installations/create-installation`)
2. Emits `controlplane.started` usage event (or this happens automatically when the cluster reports in)
3. Redirects to the deploy/progress page

**Step 3: Commit**

```bash
git add web/apps/console/src/components/kl-cloud-installation-form.tsx
git commit -m "feat(billing): rewrite installation form for pay-as-you-go with hourly pricing"
```

---

## Task 14: Rewrite Billing Settings Page

**Files:**
- Modify: `web/apps/console/src/app/installations/settings/billing/page.tsx` (157 lines)
- Rewrite: `web/apps/console/src/components/billing/subscription-management.tsx` (188 lines) → rename to `credit-management.tsx`

**Step 1: Create credit management component**

Replace `subscription-management.tsx` with `credit-management.tsx`:

Sections:
- **Credit Balance Card** — large balance number, "Add Credits" button
- **Active Usage** — table of currently running resources with per-hour cost and running total
- **Auto Top-Up Settings** — toggle, threshold input, amount input, save button
- **Transaction History** — paginated table of credit_transactions (type, amount, description, date)
- **Monthly Invoices** — "Manage Invoices" button that opens Stripe billing portal

**Step 2: Update billing settings page**

Modify `page.tsx` to:
1. Fetch credit account instead of billing account
2. Fetch active usage periods
3. Fetch pricing tiers
4. Render `<CreditManagement>` instead of `<SubscriptionManagement>`

**Step 3: Commit**

```bash
git add web/apps/console/src/components/billing/credit-management.tsx
git add web/apps/console/src/app/installations/settings/billing/page.tsx
git commit -m "feat(billing): replace subscription management with credit balance management UI"
```

---

## Task 15: Replace use-subscription-payments Hook

**Files:**
- Rewrite: `web/apps/console/src/hooks/use-subscription-payments.ts` (116 lines) → rename to `use-credits.ts`

**Step 1: Write the credits hook**

```typescript
// Returns: { loading, creditAccount, transactions, activePeriods, refresh, handleTopup, handleManageBilling, handleUpdateAutoTopup }

export function useCredits(orgId: string) {
  // Fetches GET /api/orgs/${orgId}/billing/credits on mount
  // handleTopup(amount) → POST /api/orgs/${orgId}/billing/topup → redirect to Stripe invoice URL
  // handleManageBilling() → POST /api/orgs/${orgId}/billing/portal → redirect to Stripe portal
  // handleUpdateAutoTopup(enabled, threshold, amount) → PATCH /api/orgs/${orgId}/billing/credits
  // refresh() → refetch credit data
}
```

**Step 2: Commit**

```bash
git add web/apps/console/src/hooks/use-credits.ts
git commit -m "feat(billing): add useCredits hook replacing useSubscriptionPayments"
```

---

## Task 16: Update Checkout Route for Top-Up Flow

**Files:**
- Modify: `web/apps/console/src/app/api/orgs/[orgId]/billing/checkout/route.ts` (161 lines)

**Step 1: Rewrite for credit top-up**

The checkout route currently creates Stripe Checkout sessions for subscriptions. Rewrite to:

1. Accept `{ amount: number, installationId?: string }` (the amount to top up)
2. Create or reuse Stripe customer
3. Create Stripe Invoice with the amount
4. Store `installationId` in invoice metadata if provided (for post-payment redirect)
5. Return `{ url: invoice.hosted_invoice_url }`

This is essentially the same as the top-up route (Task 6). Consider whether to merge them or keep checkout as the installation-flow-specific version that also accepts `installationId` for redirect.

**Decision:** Merge. The top-up route (Task 6) handles both cases. Remove the checkout route and point the installation form to the top-up route. The `installationId` is passed as a query parameter to the success redirect URL.

**Step 2: Update or delete the old checkout route**

Delete `web/apps/console/src/app/api/orgs/[orgId]/billing/checkout/route.ts` since the top-up route replaces it.

**Step 3: Commit**

```bash
git rm web/apps/console/src/app/api/orgs/[orgId]/billing/checkout/route.ts
git commit -m "refactor(billing): remove subscription checkout route, replaced by credit topup"
```

---

## Task 17: Update Installation Continue Route

**Files:**
- Modify: `web/apps/console/src/app/api/installations/[id]/continue/route.ts`

**Step 1: Update the continue route**

Currently checks for active subscription. Change to check for sufficient credit balance instead:

1. Look up installation → get org_id
2. Fetch credit_account for org
3. If balance > 0 → continue to deploy page
4. If balance <= 0 → redirect to billing/top-up page

**Step 2: Commit**

```bash
git add web/apps/console/src/app/api/installations/[id]/continue/route.ts
git commit -m "refactor(billing): update continue route to check credit balance instead of subscription"
```

---

## Task 18: Remove Old Subscription Routes & Components

**Files:**
- Delete: `web/apps/console/src/app/api/orgs/[orgId]/billing/subscription/route.ts`
- Delete: `web/apps/console/src/app/api/installations/[id]/subscription/route.ts`
- Delete: `web/apps/console/src/components/billing/subscription-management.tsx` (after Task 14)
- Delete: `web/apps/console/src/hooks/use-subscription-payments.ts` (after Task 15)
- Modify: `web/apps/console/src/lib/console/storage/billing.ts` — keep webhook idempotency functions, remove subscription-specific CRUD
- Modify: `web/apps/console/src/lib/console/storage/billing-types.ts` — remove `SubscriptionItem` type

**Step 1: Delete old files**

```bash
git rm web/apps/console/src/app/api/orgs/[orgId]/billing/subscription/route.ts
git rm web/apps/console/src/app/api/installations/[id]/subscription/route.ts
git rm web/apps/console/src/components/billing/subscription-management.tsx
git rm web/apps/console/src/hooks/use-subscription-payments.ts
```

**Step 2: Clean up billing storage**

Keep `isWebhookEventProcessed` and `markWebhookEventProcessed` (still needed). Remove subscription/billing-account functions or mark deprecated.

**Step 3: Search for remaining references**

```bash
rg "subscription" --include="*.ts" --include="*.tsx" web/apps/console/src/
```

Fix any remaining imports or references.

**Step 4: Commit**

```bash
git add -A
git commit -m "refactor(billing): remove old subscription-based billing routes and components"
```

---

## Task 19: Go WorkMachine Controller — Usage Event Reporting

**Files:**
- Modify: `api/internal/controllers/workmachine/cloud.go` (~572 lines)
- Create: `api/internal/controllers/workmachine/usage_reporter.go`

**Step 1: Create usage reporter helper**

```go
// usage_reporter.go
package workmachine

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "go.uber.org/zap"
)

type UsageEvent struct {
    EventType    string                 `json:"event_type"`
    ResourceID   string                 `json:"resource_id"`
    ResourceType string                 `json:"resource_type"`
    Metadata     map[string]interface{} `json:"metadata,omitempty"`
    Timestamp    time.Time              `json:"timestamp"`
}

type UsageReporter struct {
    consoleBaseURL  string
    installationKey string
    httpClient      *http.Client
    logger          *zap.Logger
}

func NewUsageReporter(consoleBaseURL, installationKey string, logger *zap.Logger) *UsageReporter {
    return &UsageReporter{
        consoleBaseURL:  consoleBaseURL,
        installationKey: installationKey,
        httpClient:      &http.Client{Timeout: 10 * time.Second},
        logger:          logger,
    }
}

func (r *UsageReporter) ReportEvent(ctx context.Context, event UsageEvent) {
    body, err := json.Marshal(event)
    if err != nil {
        r.logger.Error("failed to marshal usage event", zap.Error(err))
        return
    }

    req, err := http.NewRequestWithContext(ctx, "POST", r.consoleBaseURL+"/api/installations/usage-event", bytes.NewReader(body))
    if err != nil {
        r.logger.Error("failed to create usage event request", zap.Error(err))
        return
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-installation-key", r.installationKey)

    resp, err := r.httpClient.Do(req)
    if err != nil {
        r.logger.Error("failed to send usage event", zap.String("event_type", event.EventType), zap.Error(err))
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        r.logger.Error("usage event rejected", zap.String("event_type", event.EventType), zap.Int("status", resp.StatusCode))
    }
}
```

**Step 2: Add UsageReporter to WorkMachineReconciler**

Add `usageReporter *UsageReporter` field to the reconciler struct. Initialize in the constructor or setup.

**Step 3: Emit events from cloud.go**

In `startMachine()` (after successful start):
```go
r.usageReporter.ReportEvent(ctx, UsageEvent{
    EventType:    "workmachine.started",
    ResourceID:   machineID,
    ResourceType: wm.Spec.MachineType,
    Timestamp:    time.Now(),
})
```

In `stopMachineGracefully()` (after successful stop):
```go
r.usageReporter.ReportEvent(ctx, UsageEvent{
    EventType:    "workmachine.stopped",
    ResourceID:   machineID,
    ResourceType: wm.Spec.MachineType,
    Timestamp:    time.Now(),
})
```

In `handleStateTransitions()` for machine type change:
```go
r.usageReporter.ReportEvent(ctx, UsageEvent{
    EventType:    "workmachine.resized",
    ResourceID:   machineID,
    ResourceType: newMachineType,
    Metadata:     map[string]interface{}{"old_type": oldType, "new_type": newType},
    Timestamp:    time.Now(),
})
```

In `cleanupCloudMachine()` (deletion):
```go
r.usageReporter.ReportEvent(ctx, UsageEvent{
    EventType:    "workmachine.stopped",
    ResourceID:   machineID,
    ResourceType: wm.Spec.MachineType,
    Timestamp:    time.Now(),
})
```

**Step 4: Commit**

```bash
git add api/internal/controllers/workmachine/usage_reporter.go
git add api/internal/controllers/workmachine/cloud.go
git commit -m "feat(billing): add usage event reporting to WorkMachine controller"
```

---

## Task 20: Go kli — Extend verify-key Payload

**Files:**
- Modify: `api/cmd/kli/internal/k8s/verify.go` (108 lines)

**Step 1: Extend the request payload**

After the existing `VerifyInstallation` function, or within it, add logic to:
1. Query the Kubernetes API for running WorkMachine CRDs: `kubectl get workmachines -A -o json`
2. Query for PersistentVolumes or a storage CRD
3. Build `running_machines` and `volumes` arrays
4. Include them in the POST body to `/api/installations/verify-key`

```go
type VerifyInstallationRequest struct {
    InstallationKey string           `json:"installationKey"`
    Provider        string           `json:"provider,omitempty"`
    Region          string           `json:"region,omitempty"`
    RunningMachines []RunningMachine `json:"running_machines,omitempty"`
    Volumes         []Volume         `json:"volumes,omitempty"`
}

type RunningMachine struct {
    MachineID   string `json:"machine_id"`
    MachineType string `json:"machine_type"`
    StartedAt   string `json:"started_at"`
}

type Volume struct {
    VolumeID   string `json:"volume_id"`
    VolumeType string `json:"volume_type"` // "vm" or "object"
    SizeGB     int    `json:"size_gb"`
    CreatedAt  string `json:"created_at"`
}
```

**Step 2: Commit**

```bash
git add api/cmd/kli/internal/k8s/verify.go
git commit -m "feat(billing): extend verify-key payload with running machines and volumes"
```

---

## Task 21: Create Credit Account on Org Creation

**Files:**
- Modify: `web/apps/console/src/lib/console/storage/organizations.ts` or the org creation API route

**Step 1: Find the org creation flow**

Locate where `create_organization_with_owner()` RPC is called. After org creation, call `ensureCreditAccount(orgId)` to create the credit_accounts row with balance = 0.

**Step 2: Commit**

```bash
git add <modified files>
git commit -m "feat(billing): create credit account automatically on org creation"
```

---

## Task 22: Update E2E Tests

**Files:**
- Modify: `e2e-tests/tests/console/billing/kl-cloud-flow.test.ts`

**Step 1: Update billing flow tests**

The existing 5 billing tests follow the subscription flow. Update them for credits:

1. **Test: Add credits** — Navigate to billing settings, click "Add Credits", fill amount, complete Stripe payment, verify balance updated
2. **Test: Create installation with sufficient balance** — Add credits first, then create installation, verify no top-up prompt, verify installation created
3. **Test: Create installation with insufficient balance** — Start with $0 balance, attempt to create, verify top-up prompt shown with minimum amount
4. **Test: Auto top-up configuration** — Navigate to billing settings, enable auto top-up, set threshold and amount, save, verify settings persisted
5. **Test: View transaction history** — After top-up and usage, verify transactions appear in history

**Step 2: Run tests**

```bash
cd e2e-tests && bun run test:console
```

**Step 3: Commit**

```bash
git add e2e-tests/tests/console/billing/kl-cloud-flow.test.ts
git commit -m "test(billing): update E2E tests for pay-as-you-go credits flow"
```

---

## Task 23: Cleanup & Final Verification

**Files:**
- Various

**Step 1: Search for dead references**

```bash
rg "subscription" --include="*.ts" --include="*.tsx" web/apps/console/src/
rg "SubscriptionItem" --include="*.ts" --include="*.tsx" web/apps/console/src/
rg "billingStatus" --include="*.ts" --include="*.tsx" web/apps/console/src/
rg "currentPeriodEnd" --include="*.ts" --include="*.tsx" web/apps/console/src/
```

Fix any remaining references.

**Step 2: Build check**

```bash
cd web && bun run build:console
```

**Step 3: Run all tests**

```bash
cd e2e-tests && bun run test:console
```

**Step 4: Commit any cleanup**

```bash
git add -A
git commit -m "chore(billing): clean up remaining subscription references"
```

---

## Dependency Graph

```
Task 1 (migration) ──┬── Task 2 (types) ── Task 3 (storage CRUD) ──┬── Task 4 (usage-event route)
                      │                                              ├── Task 5 (verify-key extension)
                      │                                              ├── Task 6 (top-up route)
                      │                                              ├── Task 7 (credits route)
                      │                                              ├── Task 8 (pricing route)
                      │                                              ├── Task 9 (balance checker cron)
                      │                                              ├── Task 10 (monthly invoice cron)
                      │                                              └── Task 11 (webhook handler)
                      │
                      └── Task 12 (stripe bootstrap) ── Task 13 (installation form) ── Task 16 (checkout route)
                                                                                     ── Task 17 (continue route)

Task 14 (billing settings page) ── depends on Task 3, Task 7
Task 15 (credits hook) ── depends on Task 7
Task 18 (remove old code) ── depends on Tasks 13-17
Task 19 (Go controller) ── independent (can parallel with web tasks)
Task 20 (Go kli) ── independent (can parallel with web tasks)
Task 21 (org creation) ── depends on Task 3
Task 22 (E2E tests) ── depends on Tasks 13, 14
Task 23 (cleanup) ── depends on all above
```

**Parallelizable groups:**
- Group A (web foundation): Tasks 1 → 2 → 3 (sequential)
- Group B (API routes): Tasks 4, 5, 6, 7, 8, 9, 10, 11 (parallel after Group A)
- Group C (UI): Tasks 12, 13, 14, 15 (parallel after Group A)
- Group D (Go): Tasks 19, 20 (parallel with everything, independent)
- Group E (integration): Tasks 16, 17, 18 (after B+C)
- Group F (testing): Tasks 21, 22, 23 (after E)
