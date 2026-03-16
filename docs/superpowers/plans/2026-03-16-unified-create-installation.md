# Unified Create Installation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Merge two separate installation creation pages into one unified page with shared fields and type tabs.

**Architecture:** Extract shared form fields (name, domain) into a reusable component. Create a parent page that composes shared fields + tab content (KL Cloud billing or BYOC info). Both tabs share the same react-hook-form instance.

**Tech Stack:** Next.js 16, React 19, react-hook-form, zod, Radix Tabs, `@kloudlite/ui`

**Spec:** `docs/superpowers/specs/2026-03-16-unified-create-installation.md`

---

## Chunk 1: Create new components and page

### Task 1: Create InstallationFields component

**Files:**
- Create: `web/apps/console/src/components/installation-fields.tsx`

Extract Name + Domain fields from `kl-cloud-installation-form.tsx` (lines 219-297). This component receives `control` and `creating` as props, plus the subdomain check state. No Description field.

- [ ] **Step 1:** Create `installation-fields.tsx` with Name and Domain FormFields
- [ ] **Step 2:** Verify it compiles (no runtime test yet — composed in Task 3)
- [ ] **Step 3:** Commit

### Task 2: Create KlCloudTab and ByocTab components

**Files:**
- Create: `web/apps/console/src/components/kl-cloud-tab.tsx`
- Create: `web/apps/console/src/components/byoc-tab.tsx`

**KlCloudTab:** Extract billing section from `kl-cloud-installation-form.tsx` (lines 301-534). Receives `orgId`, `creating`, `subdomainAvailable`, `onSubmit` as props. Contains pricing, calculator, balance gate, and submit button.

**ByocTab:** Simple component with requirements list + submit button. Receives `creating`, `subdomainAvailable`, `onSubmit` as props.

- [ ] **Step 1:** Create `kl-cloud-tab.tsx` with billing section
- [ ] **Step 2:** Create `byoc-tab.tsx` with requirements info + continue button
- [ ] **Step 3:** Commit

### Task 3: Create unified page

**Files:**
- Create: `web/apps/console/src/components/create-installation-page.tsx`
- Modify: `web/apps/console/src/app/(authenticated)/installations/new/page.tsx`

**`create-installation-page.tsx`:** Client component that:
- Creates `useForm` with shared schema (name + subdomain)
- Renders `InstallationFields` at top
- Renders pill-style Tabs (Kloudlite Cloud | BYOC) below
- Each tab has its own submit handler calling the appropriate API endpoint
- Redirects to `/installations/[id]/install` after creation

**`new/page.tsx`:** Server component that fetches session/org and renders `CreateInstallationPage`.

- [ ] **Step 1:** Create `create-installation-page.tsx`
- [ ] **Step 2:** Update `new/page.tsx` to render it
- [ ] **Step 3:** Verify in browser
- [ ] **Step 4:** Commit

---

## Chunk 2: Post-creation page and cleanup

### Task 4: Create post-creation install page

**Files:**
- Create: `web/apps/console/src/app/(authenticated)/installations/[id]/install/page.tsx`

Server component that:
- Fetches installation by ID
- If `cloudProvider === 'oci'`: renders `CompletionStatus`
- Else: renders `InstallCommands`

Note: This page needs a simpler layout than the full `[id]` layout. For now, render within the existing `[id]` layout — the install content works fine there.

- [ ] **Step 1:** Create the page
- [ ] **Step 2:** Test by creating an installation and verifying redirect
- [ ] **Step 3:** Commit

### Task 5: Update routes and clean up

**Files:**
- Modify: `web/apps/console/src/components/new-installation-button.tsx` — link to `/installations/new`
- Modify: `web/apps/console/src/app/api/installations/[id]/continue/route.ts` — update redirects
- Delete: `web/apps/console/src/app/(authenticated)/installations/new-kl-cloud/` directory
- Delete: `web/apps/console/src/app/(authenticated)/installations/new-byoc/` directory
- Delete: `web/apps/console/src/components/kl-cloud-installation-form.tsx`
- Delete: `web/apps/console/src/components/installation-form.tsx`

- [ ] **Step 1:** Update `new-installation-button.tsx` link
- [ ] **Step 2:** Update continue route redirects
- [ ] **Step 3:** Delete old directories and components
- [ ] **Step 4:** Verify no broken imports
- [ ] **Step 5:** Commit

### Task 6: Final verification

- [ ] **Step 1:** Test KL Cloud flow end-to-end
- [ ] **Step 2:** Test BYOC flow end-to-end
- [ ] **Step 3:** Verify old routes are gone
