# Unified Create Installation Page

## Problem

The create installation flow is split across two separate pages (`/installations/new-kl-cloud` and `/installations/new-byoc`) with duplicated name and domain fields. This should be a single page with shared fields and tabs to select the installation type.

## Design

### Route: `/installations/new`

Single page with:
1. **Shared fields** (top) — Name + Domain in a card with "Installation Details" header. No Description field (KL Cloud doesn't have it, keep it simple).
2. **Type tabs** (below) — Pill-style tabs using `@kloudlite/ui` Tabs: "Kloudlite Cloud" | "Bring your own Cloud"
3. **Tab content:**
   - **Kloudlite Cloud:** Billing/pricing section + "Create Installation" button. Posts to `/api/installations/create-kl-installation`.
   - **BYOC:** Info about requirements (cloud CLI, account access) + "Continue to Setup" button. Posts to `/api/installations/create-installation`.

Both tabs submit the same shared fields (name, domain) and create an installation record. After creation, redirect to `/installations/[id]/install`.

### Route: `/installations/[id]/install`

Post-creation page that checks the installation type and renders:
- **KL Cloud (OCI):** Auto-deploy progress using existing `CompletionStatus` component
- **BYOC:** CLI install commands using existing `InstallCommands` component

**Layout note:** This page should NOT inherit the full `[id]` layout (which shows installation header, metadata, and detail tabs). It needs its own simple layout with just a back link. Place it under a route group or use a separate layout that opts out of the `[id]` chrome.

### Component Structure

**`create-installation-page.tsx`** — client component, parent form owner:
- Creates `useForm` with shared schema (name + subdomain)
- Wraps children in `FormProvider`
- Renders `InstallationFields` + Tabs

**`installation-fields.tsx`** — shared Name + Domain fields:
- Receives `form.control` as a prop (explicit, no context magic)
- Includes subdomain availability check via `useSubdomainCheck` hook

**`kl-cloud-tab.tsx`** — billing section:
- Pricing display, cost calculator, balance check
- "Create Installation" button that submits shared fields + creates with `cloudProvider: 'oci'`
- Extracted from `kl-cloud-installation-form.tsx` (billing logic only)

**`byoc-tab.tsx`** — requirements info:
- List of what's needed (cloud CLI, account access)
- "Continue to Setup" button that submits shared fields + creates installation

### Form Behavior

- Shared fields (name, domain) are controlled by a parent `react-hook-form` form via `FormProvider`
- `installation-fields.tsx` uses prop-drilled `control` from the parent form
- Tab switch does NOT reset the shared fields
- Validation runs on shared fields regardless of active tab
- Submit button is in each tab (different labels, different API endpoints)

## Files

### Create
- `(authenticated)/installations/new/page.tsx` — server component, fetches org/session
- `components/create-installation-page.tsx` — client component with form + tabs
- `components/installation-fields.tsx` — shared Name + Domain fields
- `components/kl-cloud-tab.tsx` — billing section for KL Cloud
- `components/byoc-tab.tsx` — requirements info for BYOC
- `(authenticated)/installations/[id]/install/page.tsx` — post-creation status page
- `(authenticated)/installations/[id]/install/layout.tsx` — simple layout (back link only, opts out of `[id]` chrome)

### Delete
- `(authenticated)/installations/new-kl-cloud/` — entire directory (page + layout)
- `(authenticated)/installations/new-byoc/` — entire directory (page + layout + content component)
- `components/installation-form.tsx` — replaced by `installation-fields.tsx` + `byoc-tab.tsx`
- `components/kl-cloud-installation-form.tsx` — replaced by `installation-fields.tsx` + `kl-cloud-tab.tsx`

### Modify
- `components/new-installation-button.tsx` — update link from `/installations/new-kl-cloud` to `/installations/new`
- `api/installations/[id]/continue/route.ts` — update hardcoded redirects from old routes to new routes
- `(authenticated)/installations/new/layout.tsx` — keep as-is (back link + max-w-6xl)

### Keep (existing sub-routes under `/installations/new/`)
- `new/install/page.tsx` — BYOC install commands (keep until migrated to `[id]/install`)
- `new/domain/page.tsx` — domain configuration
- `new/complete/page.tsx` — completion status
- `new/kloudlite-cloud/page.tsx` — KL Cloud deployment progress (keep until migrated to `[id]/install`)

### Reuse existing
- `components/completion-status.tsx` — for KL Cloud post-creation
- `components/install-commands.tsx` — for BYOC post-creation
- `hooks/use-subdomain-check.ts` — subdomain validation
- `hooks/use-credits.ts` — credit balance for billing

## Out of Scope
- Migrating the existing post-creation sub-routes (`new/install`, `new/complete`, `new/kloudlite-cloud`) to `[id]/install` — that's a follow-up
- Changing the billing/pricing logic
- Changing the installation API endpoints
