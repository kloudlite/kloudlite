# Admin-only Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert `web/apps/dashboard` into an admin-only application backed by the current `platform-api` default URL.

**Architecture:** Remove the user-facing Next route group, add a root redirect into `/admin`, and trim admin chrome links that point back to the removed user dashboard. Keep auth/admin routes and existing admin service abstractions intact, then let TypeScript/build reveal any stale imports.

**Tech Stack:** Next.js app router, React 19, TypeScript, Bun, Vitest, ESLint.

---

## File structure

- Create `web/apps/dashboard/src/app/page.tsx`: root route that redirects to `/admin`.
- Delete `web/apps/dashboard/src/app/(main)/**`: user-facing dashboard/workspace/environment/artifact/code-analysis pages and their local components.
- Modify `web/apps/dashboard/src/lib/env.ts`: change dev/build fallback API URL and warning from `http://localhost:8080` to `https://localhost:9443`.
- Modify `web/apps/dashboard/src/app/admin/layout.tsx`: stop passing removed user-dashboard affordance props to the admin profile dropdown.
- Modify `web/apps/dashboard/src/app/admin/_components/admin-profile-dropdown.tsx`: remove the “Dashboard” link to `/` and unused `hasUserRole` prop/imports.
- Modify `web/apps/dashboard/src/app/admin/_components/admin-machine-detail.tsx`: change stale back navigation from `/administration` to `/admin/machine-configs` if the component is still compiled.
- Review `web/apps/dashboard/src/app/api/**`: keep auth/admin-required API routes; delete user-only proxy routes only if no retained code imports or calls them during verification.

## Task 1: Add root admin redirect and platform-api fallback

**Files:**
- Create: `web/apps/dashboard/src/app/page.tsx`
- Modify: `web/apps/dashboard/src/lib/env.ts`

- [ ] **Step 1: Add root route redirect**

Create `web/apps/dashboard/src/app/page.tsx`:

```tsx
import { redirect } from 'next/navigation'

export default function HomePage() {
  redirect('/admin')
}
```

- [ ] **Step 2: Update the default API URL**

In `web/apps/dashboard/src/lib/env.ts`, replace the development fallback warning and return value with `https://localhost:9443`:

```ts
console.warn('⚠️  NEXT_PUBLIC_API_URL is not set. Falling back to https://localhost:9443')
```

```ts
apiUrl: apiUrl || 'https://localhost:9443',
```

- [ ] **Step 3: Run focused check**

Run from `web/apps/dashboard`:

```bash
bun run test:run
```

Expected: existing tests pass, or failures are unrelated to these two files and are captured before proceeding.

- [ ] **Step 4: Commit**

```bash
git add web/apps/dashboard/src/app/page.tsx web/apps/dashboard/src/lib/env.ts
git commit -m "feat: route dashboard root to admin"
```

## Task 2: Remove user-facing route group

**Files:**
- Delete: `web/apps/dashboard/src/app/(main)/**`

- [ ] **Step 1: Delete the route group**

Remove the complete directory:

```bash
rm -rf 'web/apps/dashboard/src/app/(main)'
```

- [ ] **Step 2: Check for stale imports into deleted route group**

Run from repository root:

```bash
rg '\(main\)|/workspaces|/environments|/artifacts|/code-analysis' web/apps/dashboard/src
```

Expected: no imports from `src/app/(main)`. URL string references may remain only in user-only API routes or dead user services that will be removed in Task 4 if they break verification.

- [ ] **Step 3: Run TypeScript/build check**

Run from `web/apps/dashboard`:

```bash
bun run build
```

Expected: build reaches compilation without missing module errors for deleted route files. If it reports stale imports, remove or update those import sites before committing.

- [ ] **Step 4: Commit**

```bash
git add -A web/apps/dashboard/src/app
git commit -m "refactor: remove user dashboard routes"
```

## Task 3: Trim admin chrome links to removed user dashboard

**Files:**
- Modify: `web/apps/dashboard/src/app/admin/layout.tsx`
- Modify: `web/apps/dashboard/src/app/admin/_components/admin-profile-dropdown.tsx`
- Modify: `web/apps/dashboard/src/app/admin/_components/admin-machine-detail.tsx`

- [ ] **Step 1: Remove user-role dashboard prop from layout**

In `web/apps/dashboard/src/app/admin/layout.tsx`, remove `hasUserRole` and pass only identity fields:

```tsx
const userRoles = session.user?.roles || []
const hasAdminRole = userRoles.includes('admin') || userRoles.includes('super-admin')
const isSuperAdmin = userRoles.includes('super-admin')
```

```tsx
<AdminProfileDropdown name={session.user?.name} email={session.user?.email} />
```

- [ ] **Step 2: Remove dashboard menu item**

In `web/apps/dashboard/src/app/admin/_components/admin-profile-dropdown.tsx`, remove `Link`, `Home`, and `hasUserRole`. The final prop interface should be:

```tsx
interface AdminProfileDropdownProps {
  name?: string | null
  email?: string | null
}
```

The component signature should be:

```tsx
export function AdminProfileDropdown({ name, email }: AdminProfileDropdownProps) {
```

The dropdown should contain the account label, separator, and sign-out item only.

- [ ] **Step 3: Fix stale admin machine back link**

In `web/apps/dashboard/src/app/admin/_components/admin-machine-detail.tsx`, replace:

```tsx
onClick={() => router.push('/administration')}
```

with:

```tsx
onClick={() => router.push('/admin/machine-configs')}
```

- [ ] **Step 4: Run lint**

Run from `web/apps/dashboard`:

```bash
bun run lint
```

Expected: no unused import/prop errors from admin chrome changes.

- [ ] **Step 5: Commit**

```bash
git add web/apps/dashboard/src/app/admin/layout.tsx web/apps/dashboard/src/app/admin/_components/admin-profile-dropdown.tsx web/apps/dashboard/src/app/admin/_components/admin-machine-detail.tsx
git commit -m "refactor: trim admin-only navigation"
```

## Task 4: Remove user-only proxies only if verification proves they are dead

**Files:**
- Candidate delete: `web/apps/dashboard/src/app/api/vscode/**`
- Candidate delete: `web/apps/dashboard/src/app/api/v1/namespaces/[namespace]/services/[name]/logs/route.ts`
- Candidate delete: `web/apps/dashboard/src/app/api/v1/work-machines/[name]/metrics/route.ts`

- [ ] **Step 1: Search for retained callers**

Run from repository root:

```bash
rg '/api/vscode|/api/v1/namespaces/.*/services/.*/logs|/api/v1/work-machines/.*/metrics' web/apps/dashboard/src
```

Expected: no retained admin/auth components call these endpoints after deleting `(main)`.

- [ ] **Step 2: Delete uncalled user-only proxies**

If Step 1 only finds the route files themselves, delete the candidate route files/directories. Keep unrelated API routes such as auth, superadmin login, health, download, vpn, and resource-events.

- [ ] **Step 3: Run build**

Run from `web/apps/dashboard`:

```bash
bun run build
```

Expected: build succeeds or reports only pre-existing runtime/config issues unrelated to deleted user routes. Fix missing imports caused by this task before committing.

- [ ] **Step 4: Commit if files changed**

```bash
git add -A web/apps/dashboard/src/app/api
git commit -m "refactor: remove user dashboard api proxies"
```

If no candidate files were deleted, skip this commit.

## Task 5: Final verification and review

**Files:**
- Verify all changed dashboard files.

- [ ] **Step 1: Run dashboard tests**

Run from `web/apps/dashboard`:

```bash
bun run test:run
```

Expected: Vitest passes.

- [ ] **Step 2: Run dashboard lint**

Run from `web/apps/dashboard`:

```bash
bun run lint
```

Expected: ESLint passes.

- [ ] **Step 3: Run dashboard build**

Run from `web/apps/dashboard`:

```bash
bun run build
```

Expected: Next build passes, or any blocker is documented with exact output and judged unrelated before stopping.

- [ ] **Step 4: Inspect final diff**

Run from repository root:

```bash
git status --short
git diff --stat HEAD~4..HEAD
```

Expected: changed files are limited to the admin-only dashboard cleanup and docs/spec/plan commits.

- [ ] **Step 5: Request review before PR/integration**

Dispatch the configured `reviewer` agent with the branch diff and ask for correctness, scope, verification, CI risk, and deployment/runtime risk review.

---

## Self-review

- Spec coverage: root redirects to `/admin`; user routes are deleted; admin routes/auth remain; default API URL points at `https://localhost:9443`; admin links to removed dashboard are trimmed; verification includes tests, lint, and build.
- Placeholder scan: no open-ended implementation placeholders are required for workers; conditional API-proxy deletion has exact criteria and commands.
- Type consistency: `AdminProfileDropdown` prop removal is reflected in both `admin/layout.tsx` and the component definition.
