# Pay-As-You-Go Prepaid Credits Billing Design

**Date:** 2026-03-09
**Status:** Approved
**Phase:** 6 of the billing migration (after Razorpay→Stripe, server action→API route, schema rename, PII separation, org-level restructuring)

## Overview

Replace the subscription-based billing model with pay-as-you-go prepaid credits. Users purchase credits upfront, and usage is metered hourly and deducted from their balance. Stripe is used only for payment collection (invoices) and informational monthly summaries — no Stripe subscriptions or metered billing API.

## Billable Units

| Resource | Rate |
|----------|------|
| Control Plane (master node) | $0.02/hr |
| WorkMachine VM (by tier/size) | varies/hr (e.g., Standard $0.05/hr, Performance $0.12/hr) |
| Storage — VM Volume | $0.04/GB/mo (~$0.0000556/GB/hr) |
| Storage — Object Storage | $0.04/GB/mo (~$0.0000556/GB/hr) |

## Architecture

### Credit Ledger (our DB, not Stripe)

All balance tracking lives in our Supabase/PostgreSQL database. Stripe handles payment collection only.

- **Top-up flow:** User requests credits → Stripe Invoice (mode: `payment`) → webhook/polling confirms payment → credits added to our `credit_accounts` table
- **Usage deduction:** Cron job runs every 5 minutes, calculates cost of active resources since last calculation, debits from credit balance
- **Monthly invoice:** Cron job on the 1st of each month creates a Stripe Invoice (informational, $0 due) showing the usage breakdown for the previous period

### Communication Channels

```
[In-Cluster Controller] --HTTP POST /api/installations/usage-event--> [Console API]
[In-Cluster kli]        --HTTP POST /api/installations/verify-key---> [Console API] (heartbeat every ~10 min)
[Console API]           --Stripe SDK--> [Stripe] (invoices only)
[Stripe Webhook]        --POST /api/webhooks/stripe--> [Console API] (payment confirmations)
```

## Data Model

### credit_accounts

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | PK |
| org_id | uuid | FK to orgs |
| balance | numeric(12,4) | Current credit balance in dollars |
| auto_topup_enabled | boolean | Default false |
| auto_topup_threshold | numeric(10,2) | Trigger auto top-up when balance drops below this |
| auto_topup_amount | numeric(10,2) | Amount to top up |
| stripe_customer_id | text | Stripe customer for invoicing |
| created_at | timestamptz | |
| updated_at | timestamptz | |

One row per org. Created when the org is created.

### credit_transactions

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | PK |
| org_id | uuid | FK |
| type | text | `topup`, `usage_debit`, `adjustment` |
| amount | numeric(12,4) | Positive for top-ups, negative for debits |
| description | text | Human-readable (e.g., "Top-up via Stripe Invoice inv_xxx") |
| stripe_invoice_id | text | Nullable, set for top-ups |
| usage_period_id | uuid | Nullable, set for usage debits |
| created_at | timestamptz | |

Append-only ledger. Balance in `credit_accounts` is the materialized sum.

### usage_events

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | PK |
| installation_id | uuid | FK |
| event_type | text | `workmachine.started`, `workmachine.stopped`, `workmachine.resized`, `controlplane.started`, `controlplane.stopped`, `storage.provisioned`, `storage.resized`, `storage.deleted` |
| resource_id | text | machine_id or volume_id |
| resource_type | text | Machine type/tier, or `vm_volume`/`object_storage` |
| metadata | jsonb | Additional data (size_gb, old_type, new_type, etc.) |
| timestamp | timestamptz | When the event occurred |
| created_at | timestamptz | When we received it |

Unique constraint on `(installation_id, resource_id, event_type, timestamp)` to prevent duplicates.

### usage_periods

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | PK |
| installation_id | uuid | FK |
| org_id | uuid | FK (denormalized for efficient queries) |
| resource_id | text | machine_id or volume_id |
| resource_type | text | Machine tier or volume type |
| started_at | timestamptz | |
| ended_at | timestamptz | Null while active |
| hourly_rate | numeric(10,6) | Rate at time of start |
| total_cost | numeric(12,4) | Calculated when ended, or on periodic debit |
| last_billed_at | timestamptz | Last time cost was deducted from credits |

Open periods (ended_at IS NULL) represent currently running resources.

## Key Flows

### 1. Installation Creation

1. User selects region and WorkMachine configuration
2. Console calculates 30-day projected cost: `(CP hourly rate + WM hourly rate + storage rate) × 24 × 30`
3. If `credit_accounts.balance >= projected_30day_cost` → proceed to deploy
4. If balance is insufficient → show top-up page with minimum = `projected_30day_cost - current_balance`
5. Top-up creates a Stripe Invoice (mode: `payment`) with the amount
6. On payment confirmation (webhook or polling) → add credits → redirect to deploy

### 2. Usage Event Processing

1. WorkMachine controller detects VM state change → POST to `/api/installations/usage-event`
2. Console validates installation key, inserts into `usage_events`
3. For `started`/`provisioned` events → creates new `usage_period` (open)
4. For `stopped`/`deleted` events → closes the matching `usage_period`, calculates cost, creates `credit_transaction` debit

### 3. Balance Checker Cron (every 5 minutes)

1. Query all open `usage_periods`
2. For each, calculate cost accrued since `last_billed_at`
3. Group by org, debit total from `credit_accounts.balance` atomically
4. Update `last_billed_at` on each usage period
5. Check projected time until $0:
   - < 24 hours → trigger low-balance warning (set flag for UI banner)
   - <= $0 → pause all WorkMachines (close usage periods), keep control plane running
   - < -$10 → pause control plane too, flag org for review
6. If auto top-up enabled and balance < threshold → create Stripe Invoice for auto top-up amount

### 4. Heartbeat Reconciliation (on each verify-key call)

1. `kli` sends current state: `running_machines` and `volumes` arrays
2. Console compares against open `usage_periods` for this installation
3. Machine/volume running but no open period → create one (missed `started` event)
4. Open period exists but resource not in heartbeat → close it (missed `stopped` event)
5. Update `installations.last_health_check`

### 5. Monthly Invoice Generation (1st of each month)

1. Query all `credit_transactions` of type `usage_debit` for the previous month, grouped by org
2. For each org, create a Stripe Invoice (informational, $0 due) with line items:
   - Control Plane hours × rate
   - WorkMachine hours × rate (per machine/tier)
   - VM Storage GB-hours × rate
   - Object Storage GB-hours × rate
3. Send via Stripe (email receipt)

## UX Changes

### Installation Creation Form

Replace subscription tier selection with:

1. **Choose region** (unchanged)
2. **Choose WorkMachine configuration** — cards showing VM sizes with hourly rates
3. **Cost summary panel** — projected monthly cost assuming 24/7 uptime:
   - Control Plane: ~$14.40/mo
   - 1× Standard WorkMachine: ~$36.00/mo
   - Storage (50 GB): ~$2.00/mo
   - Estimated total: ~$52.40/mo
   - Disclaimer: "Billed by actual usage. WorkMachines only charged while running."
4. **Balance gate:** sufficient balance → "Create Installation" button; insufficient → "Add Credits" with calculated minimum

### Billing Settings Page

- **Credit Balance** — large display of current balance
- **Usage This Period** — breakdown by resource type
- **Top-Up Credits** — manual top-up with minimum shown
- **Auto Top-Up** — toggle + threshold + amount config
- **Transaction History** — credit additions and usage debits
- **Monthly Invoices** — list of Stripe informational invoices

### Header

- Optional: small credit balance indicator next to settings gear (nice-to-have for v2)

## Go-Side Agent Changes

### WorkMachine Controller

Add HTTP POST to console API on VM state changes:

| Event | Trigger |
|-------|---------|
| `workmachine.started` | VM provisioned and running |
| `workmachine.stopped` | VM stopped |
| `workmachine.resized` | VM tier changed |
| `controlplane.started` | Master node comes up |
| `controlplane.stopped` | Master node torn down |
| `storage.provisioned` | Volume created (vm or object type) |
| `storage.resized` | Volume expanded |
| `storage.deleted` | Volume deleted |

**API route:** `POST /api/installations/usage-event`
- Auth: `x-installation-key` header
- Body: `{ event_type, resource_id, resource_type, metadata, timestamp }`
- Fire-and-forget from controller (log errors, don't block reconciliation)

### verify-key Extension

Add to request body:
```json
{
  "key": "...",
  "running_machines": [
    { "machine_id": "wm-abc", "machine_type": "standard", "started_at": "..." }
  ],
  "volumes": [
    { "volume_id": "pv-abc", "volume_type": "vm", "size_gb": 50, "created_at": "..." },
    { "volume_id": "obj-def", "volume_type": "object", "size_gb": 100, "created_at": "..." }
  ]
}
```

## Error Handling & Edge Cases

### Credit Exhaustion

1. **Low balance** (projected $0 within 24 hours) — UI banner warning. Auto top-up triggers if enabled.
2. **Zero balance** — WorkMachines paused (not deleted). Control plane stays running. Storage continues accruing (can't delete user data), balance goes negative.
3. **Negative balance cap** — at -$10, pause control plane too and flag org for review.
4. **Credits added after exhaustion** — WorkMachines resume automatically if balance covers at least 1 hour of projected cost.

### Missed Events

- **Controller restart** — next heartbeat (~10 min) reconciles missing start/stop events
- **Network partition** — same heartbeat reconciliation catches it
- **Duplicate events** — unique constraint on usage_events rejects silently

### Installation Deletion

1. Close all open usage periods
2. Calculate and debit final charges
3. Trigger OCI uninstall job
4. Delete installation record after job completes
5. Credit account persists at org level — balance carries forward

### Stripe Failures

- **Payment fails** — credits not added, user sees error, can retry
- **Stripe outage** — top-up unavailable, existing credits continue working, banner shown
- **Monthly invoice creation fails** — log error, retry next cron run (non-blocking)

### Race Conditions

- **Concurrent top-ups** — atomic `UPDATE credit_accounts SET balance = balance + $amount`
- **Balance check vs. debit** — acceptable slight negative, corrected on next top-up
- **Multiple installations** — all share org credit account, atomic balance updates
