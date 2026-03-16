# Console Design Consistency Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Standardize all console app pages to use consistent typography, content widths, card patterns, tabs, navigation, and danger zones.

**Architecture:** Pure CSS/class changes across ~12 files. No new dependencies. One new component (`FilterTabs`). One component deletion (`InstallationSettingsTabs`). All changes are visual — no logic changes.

**Tech Stack:** Next.js 16, Tailwind CSS 4, React 19, `@kloudlite/ui`, `@kloudlite/lib`

**Spec:** `docs/superpowers/specs/2026-03-16-console-design-consistency.md`

---

## Chunk 1: Layout Width & Typography

### Task 1: Settings layout — width + typography

**Files:**
- Modify: `web/apps/console/src/app/installations/settings/layout.tsx`

- [ ] **Step 1: Fix max-width and typography**

Change line 32:
```tsx
<main className="mx-auto max-w-4xl px-6 lg:px-12 py-10">
```

Change lines 35-38:
```tsx
<h1 className="text-2xl font-semibold text-foreground">Settings</h1>
<p className="text-muted-foreground mt-1 text-sm">
  Manage your organization and billing
</p>
```

- [ ] **Step 2: Verify in browser**

Run: open `http://localhost:3002/installations/settings/organization`
Expected: "Settings" heading is smaller, content area is narrower, subtitle is standard size

- [ ] **Step 3: Commit**

```bash
git add web/apps/console/src/app/installations/settings/layout.tsx
git commit -m "fix(console): standardize settings layout width and typography"
```

### Task 2: Installation detail layout — width + padding

**Files:**
- Modify: `web/apps/console/src/app/installations/[id]/layout.tsx`

- [ ] **Step 1: Fix max-width and padding**

Change line 55:
```tsx
<main className="mx-auto max-w-4xl px-6 lg:px-12 py-10">
```

- [ ] **Step 2: Verify in browser**

Run: open `http://localhost:3002/installations/<any-id>`
Expected: Content area is narrower, padding matches settings page

- [ ] **Step 3: Commit**

```bash
git add web/apps/console/src/app/installations/[id]/layout.tsx
git commit -m "fix(console): standardize installation detail layout width and padding"
```

### Task 3: Profile page — full redesign

**Files:**
- Modify: `web/apps/console/src/app/profile/page.tsx`

- [ ] **Step 1: Remove back link, fix width, typography, and card wrapping**

Replace the entire return block (lines 26-148) with:

```tsx
return (
    <div className="bg-background flex h-screen flex-col">
      <InstallationsHeader
        user={session.user}
        orgs={orgs.map((o) => ({ id: o.id, name: o.name, slug: o.slug }))}
        currentOrgId={currentOrg?.id}
      />

      <ScrollArea className="flex-1">
        <main className="mx-auto max-w-4xl px-6 lg:px-12 py-10">
          {/* Title */}
          <div className="mb-8">
            <h1 className="text-2xl font-semibold text-foreground">Profile</h1>
            <p className="text-muted-foreground mt-1 text-sm">
              Your personal account information
            </p>
          </div>

          {/* Profile Content */}
          <div className="space-y-6">
            {/* Profile Information Card */}
            <div className="border border-foreground/10 rounded-lg p-6 bg-background">
              <div className="mb-5">
                <h2 className="text-lg font-semibold">Profile Information</h2>
                <p className="text-muted-foreground mt-1 text-sm">Your account details from {session.provider} OAuth</p>
              </div>

              <div className="space-y-5">
                {/* Profile Picture */}
                <div className="flex items-start gap-6">
                  <Avatar className="ring-foreground/10 h-24 w-24 ring-1">
                    <AvatarImage src={session.user.image} alt={session.user.name} />
                    <AvatarFallback className="text-2xl">{getInitials(session.user.name)}</AvatarFallback>
                  </Avatar>
                  <div className="flex-1 space-y-1 pt-2">
                    <label className="text-foreground text-sm font-medium">Profile Picture</label>
                    <p className="text-muted-foreground text-sm">
                      Synced from your {session.provider} account
                    </p>
                  </div>
                </div>

                <div className="h-px bg-foreground/10" />

                {/* Name */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <User className="text-muted-foreground h-4 w-4" />
                    <label className="text-foreground text-sm font-medium">Name</label>
                  </div>
                  <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
                    <p className="text-sm">{session.user.name}</p>
                  </div>
                  <p className="text-muted-foreground text-sm">
                    Synced from your {session.provider} account
                  </p>
                </div>

                <div className="h-px bg-foreground/10" />

                {/* Email */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Mail className="text-muted-foreground h-4 w-4" />
                    <label className="text-foreground text-sm font-medium">Email Address</label>
                  </div>
                  <div className="border-foreground/10 bg-muted/30 border px-4 py-3">
                    <p className="text-sm">{session.user.email}</p>
                  </div>
                  <p className="text-muted-foreground text-sm">
                    Primary email from your {session.provider} account
                  </p>
                </div>

                <div className="h-px bg-foreground/10" />

                {/* Authentication */}
                <div className="space-y-3">
                  <div className="flex items-center gap-2">
                    <Shield className="text-muted-foreground h-4 w-4" />
                    <label className="text-foreground text-sm font-medium">Authentication Provider</label>
                  </div>
                  <div className="border-foreground/10 bg-muted/30 flex items-center gap-3 border px-4 py-3">
                    <Badge variant="outline" className="capitalize">
                      {session.provider}
                    </Badge>
                    <span className="text-muted-foreground text-sm">
                      You&apos;re signed in with {session.provider}
                    </span>
                  </div>
                </div>
              </div>
            </div>

            {/* Info Note */}
            <div className="bg-muted/30 flex items-start gap-3 border border-foreground/10 rounded-lg p-4">
              <Info className="text-muted-foreground mt-0.5 h-4 w-4 flex-shrink-0" />
              <p className="text-muted-foreground text-sm">
                Your profile information is managed by your OAuth provider ({session.provider}). To update
                your name, email, or profile picture, please update them in your {session.provider}{' '}
                account settings.
              </p>
            </div>
          </div>
        </main>
      </ScrollArea>
    </div>
  )
```

Also remove the `ArrowLeft` and `Link` imports (no longer needed since back link is removed).

- [ ] **Step 2: Verify in browser**

Run: open `http://localhost:3002/profile`
Expected: No back link, smaller heading, content in a card, consistent width

- [ ] **Step 3: Commit**

```bash
git add web/apps/console/src/app/profile/page.tsx
git commit -m "fix(console): standardize profile page layout, remove back link, use cards"
```

### Task 4: Installations list page — width + padding

**Files:**
- Modify: `web/apps/console/src/app/installations/page.tsx`

- [ ] **Step 1: Fix max-width and padding**

Change line 60:
```tsx
<main className="mx-auto max-w-6xl px-6 lg:px-12 py-10">
```

- [ ] **Step 2: Commit**

```bash
git add web/apps/console/src/app/installations/page.tsx
git commit -m "fix(console): standardize installations list page width and padding"
```

### Task 5: Create installation shell — width + padding

**Files:**
- Modify: `web/apps/console/src/components/installation-layout.tsx`

- [ ] **Step 1: Fix max-width and padding**

Change line 15:
```tsx
<div className="mx-auto max-w-4xl px-6 lg:px-12 py-10">
```

- [ ] **Step 2: Commit**

```bash
git add web/apps/console/src/components/installation-layout.tsx
git commit -m "fix(console): standardize create installation shell width and padding"
```

---

## Chunk 2: Section Headings & Danger Zones

### Task 6: Organization settings — section headings + subtitles

**Files:**
- Modify: `web/apps/console/src/app/installations/settings/organization/page.tsx`

- [ ] **Step 1: Fix section heading and subtitle sizes**

Change all `text-xl font-semibold` to `text-lg font-semibold` for section headings.
Change all `text-base` subtitles (like "Manage who has access...") to `text-sm text-muted-foreground`.

Specifically in the return block:
- Line with `<h2 className="text-xl font-semibold">` → `<h2 className="text-lg font-semibold">`
- Line with `text-muted-foreground mt-1 text-base` → `text-muted-foreground mt-1 text-sm`

Apply to these section headings: "Team Members", "Pending Invitations", "Danger Zone".

Note: The org name heading (`{currentOrg.name}`) also uses `text-xl font-semibold` but should stay as-is — it's an identity display, not a section heading.

Also add a subtitle under the "Danger Zone" heading to match the standardized pattern:
```tsx
<h2 className="text-lg font-semibold">Danger Zone</h2>
<p className="text-muted-foreground mt-1 text-sm">
  Irreversible actions that affect your organization
</p>
```

- [ ] **Step 2: Commit**

```bash
git add web/apps/console/src/app/installations/settings/organization/page.tsx
git commit -m "fix(console): standardize organization settings headings and subtitles"
```

### Task 7: Billing page — heading + loading/empty/error states

**Files:**
- Modify: `web/apps/console/src/app/installations/settings/billing/page.tsx`

- [ ] **Step 1: Fix heading**

Change line 22:
```tsx
<h2 className="text-lg font-semibold">Billing</h2>
```

- [ ] **Step 2: Add Suspense boundary with proper fallback**

Wrap the `<CreditManagement>` component in a Suspense boundary with a skeleton loading state, and add an error boundary. Replace lines 20-26 with:

```tsx
import { Suspense } from 'react'

// In the return:
<div className="space-y-6">
  <div>
    <h2 className="text-lg font-semibold">Billing</h2>
    <p className="text-muted-foreground text-sm">Manage your credit balance and usage</p>
  </div>
  <Suspense fallback={
    <div className="border border-foreground/10 rounded-lg p-6 bg-background">
      <div className="space-y-4">
        <div className="h-6 w-48 bg-muted/50 rounded animate-pulse" />
        <div className="h-4 w-64 bg-muted/30 rounded animate-pulse" />
        <div className="h-32 bg-muted/20 rounded animate-pulse" />
      </div>
    </div>
  }>
    <CreditManagement orgId={currentOrg.id} isOwner={isOwner} />
  </Suspense>
</div>
```

- [ ] **Step 3: Commit**

```bash
git add web/apps/console/src/app/installations/settings/billing/page.tsx
git commit -m "fix(console): standardize billing heading, add loading skeleton"
```

### Task 8: Delete organization component — danger card tokens

**Files:**
- Modify: `web/apps/console/src/components/delete-organization.tsx`

- [ ] **Step 1: Normalize danger card classes**

Change line 56:
```tsx
<div className="border border-destructive/20 rounded-lg p-6 bg-destructive/[0.03]">
```

- [ ] **Step 2: Commit**

```bash
git add web/apps/console/src/components/delete-organization.tsx
git commit -m "fix(console): standardize delete organization danger card tokens"
```

### Task 9: Installation detail page — danger zone standardization

**Files:**
- Modify: `web/apps/console/src/app/installations/[id]/page.tsx`

- [ ] **Step 1: Read the file to find exact danger zone classes**

Read the installation detail page file to identify all danger zone styling that needs to change.

- [ ] **Step 2: Standardize danger zone**

Change danger zone heading from `text-xl` to `text-lg font-semibold`.
Change all raw `red-500`/`red-600` to semantic `destructive` tokens:
- `border-red-500/20` → `border-destructive/20`
- `bg-red-500/5` or `bg-red-500/[0.03]` → `bg-destructive/[0.03]`
- `text-red-600 dark:text-red-400` → `text-destructive`

Keep the `AlertTriangle` icon.

- [ ] **Step 3: Commit**

```bash
git add web/apps/console/src/app/installations/[id]/page.tsx
git commit -m "fix(console): standardize installation detail danger zone tokens"
```

---

## Chunk 3: Tabs Standardization

### Task 10: Update NavTabs — fix text size and underline position

**Files:**
- Modify: `web/apps/console/src/components/nav-tabs.tsx`

- [ ] **Step 1: Fix tab text size and underline position**

Change line 68 — tab text from `text-base` to `text-sm`:
```tsx
'relative px-6 py-2.5 text-sm font-medium transition-all duration-200 cursor-pointer flex items-center gap-2',
```

Underline is already at `bottom-0` (line 84) — no change needed.

- [ ] **Step 2: Commit**

```bash
git add web/apps/console/src/components/nav-tabs.tsx
git commit -m "fix(console): standardize NavTabs text size to text-sm"
```

### Task 11: Replace InstallationSettingsTabs with NavTabs

**Files:**
- Modify: `web/apps/console/src/app/installations/settings/layout.tsx`
- Delete: `web/apps/console/src/components/installation-settings-tabs.tsx`

- [ ] **Step 1: Update settings layout to use NavTabs**

In `web/apps/console/src/app/installations/settings/layout.tsx`:

Replace import:
```tsx
import { InstallationSettingsTabs } from '@/components/installation-settings-tabs'
```
with:
```tsx
import { NavTabs } from '@/components/nav-tabs'
```

Replace usage (line 42):
```tsx
<NavTabs tabs={[
  { id: 'organization', label: 'Organization', href: '/installations/settings/organization' },
  { id: 'billing', label: 'Billing', href: '/installations/settings/billing' },
]} />
```

- [ ] **Step 2: Delete the old component**

```bash
rm web/apps/console/src/components/installation-settings-tabs.tsx
```

- [ ] **Step 3: Verify in browser**

Run: open `http://localhost:3002/installations/settings/organization`
Expected: Tabs look the same visually, underline animates correctly between Organization and Billing

- [ ] **Step 4: Commit**

```bash
git add web/apps/console/src/app/installations/settings/layout.tsx
git add -u web/apps/console/src/components/installation-settings-tabs.tsx
git commit -m "fix(console): replace InstallationSettingsTabs with shared NavTabs"
```

### Task 12: Create FilterTabs component + migrate installations list

**Files:**
- Create: `web/apps/console/src/components/filter-tabs.tsx`
- Modify: `web/apps/console/src/components/installations-list.tsx`

- [ ] **Step 1: Create FilterTabs component**

```tsx
'use client'

import { useState, useRef, useEffect, useCallback } from 'react'
import { cn } from '@kloudlite/lib'

export interface FilterTab {
  id: string
  label: string
}

interface FilterTabsProps {
  tabs: FilterTab[]
  activeTab: string
  onTabChange: (tabId: string) => void
  className?: string
}

export function FilterTabs({ tabs, activeTab, onTabChange, className }: FilterTabsProps) {
  const [underlineStyle, setUnderlineStyle] = useState({ left: 0, width: 0 })
  const tabRefs = useRef<Map<string, HTMLButtonElement>>(new Map())

  const updatePosition = useCallback(() => {
    const activeRef = tabRefs.current.get(activeTab)
    if (activeRef) {
      const fullWidth = activeRef.offsetWidth
      const underlineWidth = fullWidth * 0.6
      const leftOffset = activeRef.offsetLeft + (fullWidth - underlineWidth) / 2
      setUnderlineStyle({ left: leftOffset, width: underlineWidth })
    }
  }, [activeTab])

  useEffect(() => {
    setTimeout(updatePosition, 10)
    window.addEventListener('resize', updatePosition)
    return () => window.removeEventListener('resize', updatePosition)
  }, [updatePosition])

  return (
    <div className={cn('inline-flex gap-1 relative', className)} role="tablist">
      {tabs.map((tab) => (
        <button
          key={tab.id}
          ref={(el) => {
            if (el) tabRefs.current.set(tab.id, el)
          }}
          role="tab"
          aria-selected={activeTab === tab.id}
          onClick={() => onTabChange(tab.id)}
          className={cn(
            'relative cursor-pointer px-5 py-2 text-sm font-medium transition-all duration-200',
            'hover:bg-foreground/[0.03] active:bg-foreground/[0.05] rounded-sm',
            activeTab === tab.id
              ? 'text-foreground'
              : 'text-muted-foreground hover:text-foreground',
          )}
        >
          {tab.label}
        </button>
      ))}

      {underlineStyle.width > 0 && (
        <div
          className="absolute bottom-0 h-[2px] bg-primary transition-all duration-300 ease-out"
          style={{
            left: `${underlineStyle.left}px`,
            width: `${underlineStyle.width}px`,
          }}
        />
      )}
    </div>
  )
}
```

- [ ] **Step 2: Update installations-list.tsx to use FilterTabs**

Add import:
```tsx
import { FilterTabs } from '@/components/filter-tabs'
```

Replace the entire "Status Filter Tabs" section (the `<div className="border-foreground/10 mb-5 border-b">` block, roughly lines 267-327) with:
```tsx
<div className="border-foreground/10 mb-5 border-b">
  <FilterTabs
    tabs={[
      { id: 'all', label: 'All' },
      { id: 'pending', label: 'Pending' },
      { id: 'installed', label: 'Installed' },
    ]}
    activeTab={statusFilter}
    onTabChange={(id) => setStatusFilter(id as 'all' | 'pending' | 'installed')}
  />
</div>
```

Remove the now-unused refs (`allRef`, `pendingRef`, `installedRef`), the `handleFilter*` callbacks, and the old `underlineStyle` state + effect that powered the inline tab underline.

- [ ] **Step 3: Verify in browser**

Run: open `http://localhost:3002/installations`
Expected: Filter tabs look the same, underline animates on click, filtering works

- [ ] **Step 4: Commit**

```bash
git add web/apps/console/src/components/filter-tabs.tsx
git add web/apps/console/src/components/installations-list.tsx
git commit -m "feat(console): extract FilterTabs component, migrate installations list"
```

---

## Chunk 4: Final Verification

### Task 13: Visual verification of all pages

- [ ] **Step 1: Check each page in browser**

Verify these pages have consistent styling:
1. `http://localhost:3002/installations` — `max-w-6xl`, `text-2xl` title, FilterTabs
2. `http://localhost:3002/installations/<id>` — `max-w-4xl`, danger zone with `destructive` tokens
3. `http://localhost:3002/installations/settings/organization` — `max-w-4xl`, `text-2xl` settings title, `text-lg` section headings, `text-sm` subtitles
4. `http://localhost:3002/installations/settings/billing` — `text-lg` heading
5. `http://localhost:3002/profile` — no back link, `max-w-4xl`, card wrapping, `text-2xl` title
6. `http://localhost:3002/installations/new-kl-cloud` — `max-w-4xl` shell
7. `http://localhost:3002/installations/new-byoc` — `max-w-4xl` shell

- [ ] **Step 2: Check dark mode**

Toggle to dark mode and verify all pages still look correct (semantic tokens handle this automatically).

- [ ] **Step 3: Final commit if any touch-ups needed**
