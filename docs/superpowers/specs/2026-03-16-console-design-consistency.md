# Console App Design Consistency Spec

## Problem

The console app has 8 design inconsistencies across pages — different heading sizes, content widths, card patterns, navigation behavior, tab implementations, and danger zone layouts. This makes the app feel unpolished and increases maintenance burden.

## Design Standards

### Typography Scale

| Role | Classes | Usage |
|------|---------|-------|
| Page title | `text-2xl font-semibold` | Installations, Settings, Profile, Create Installation |
| Page subtitle | `text-sm text-muted-foreground` | Below page title |
| Section heading | `text-lg font-semibold` | Team Members, Installation Details, Danger Zone, Billing, etc. |
| Section subtitle | `text-sm text-muted-foreground` | Below section heading |
| Field label | `text-sm font-medium` | Status, Name, Email, etc. |
| Body text | `text-sm` | Default content |
| Help text | `text-xs text-muted-foreground` | Tertiary info |

### Content Width

Two tiers:

- **Detail/form pages**: `max-w-4xl` (896px) with `px-6 lg:px-12 py-10` — Settings, Profile, Installation detail, Create installation forms
- **List/table pages**: `max-w-6xl` (1152px) with `px-6 lg:px-12 py-10` — Installations list (needs wider space for multi-column table)

### Card Patterns

Two card types, used consistently:

**Content card:**
```
border border-foreground/10 rounded-lg p-6 bg-background
```
Used for: Installation Details, Super Admin Access, Profile Information, Team Members table wrapper, Billing sections.

**Danger card:**
```
border-destructive/20 bg-destructive/[0.03] rounded-lg p-6
```
Used for: Danger Zone sections on both installation detail and organization pages. Uses semantic `destructive` token (not raw `red-500`/`red-600`). Internal layout: heading + description on left, action button on right. Multiple danger items stack as separate cards within a `space-y-4` container.

### Navigation Rules

- **Back link** (`← Back to Installations`): Only on pages nested under `/installations/[id]` and `/installations/new*`
- **Profile page**: No back link. It's a top-level page accessed from header avatar dropdown.
- **Settings page**: No back link. It's a top-level page accessed from header gear icon.

### Tabs

Two tab patterns sharing the same visual style:

- **Navigation tabs** (Organization/Billing, Overview): Use shared `NavTabs` component with `<Link>` elements for URL-based navigation. Already implemented.
- **Filter tabs** (All/Pending/Installed): Create a `FilterTabs` component that shares visual styling with `NavTabs` but uses `<button>` elements with callback-based state. The current custom inline implementation in `installations-list.tsx` should be extracted into this component.

Note: `NavTabs` uses `<Link>` for routing; filter tabs use `<button>` for local state. These are different interaction patterns that should not be forced into one component.

Shared visual standards for both tab types:
- Tab text: `text-sm font-medium`
- Underline position: `bottom-0`
- Active underline: `bg-primary h-[2px]` with 60% tab width, centered

### Danger Zone Standardization

Both installation detail and organization settings should use the same Danger Zone pattern:
- Section heading: `text-lg font-semibold` "Danger Zone"
- Section subtitle: `text-sm text-muted-foreground` "Irreversible actions..."
- Each danger item is a separate danger card with:
  - Left side: item title (`text-base font-semibold text-destructive`) + description
  - Right side: destructive action button

## Changes by File

### 1. Settings layout (`installations/settings/layout.tsx`)
- Change `max-w-7xl` to `max-w-4xl`
- Change page title from `text-4xl lg:text-5xl font-bold tracking-tight` to `text-2xl font-semibold`
- Change subtitle from `text-[1.0625rem]` to `text-sm text-muted-foreground`
- Normalize padding to `px-6 lg:px-12 py-10` (already correct)

### 2. Installation detail layout (`installations/[id]/layout.tsx`)
- Change `max-w-7xl` to `max-w-4xl`
- Normalize padding from `px-6 lg:px-8 py-8` to `px-6 lg:px-12 py-10`

### 3. Profile page (`profile/page.tsx`)
- Change outer `max-w-7xl` to `max-w-4xl` and remove inner `max-w-2xl` constraint
- Remove "Back to Installations" link
- Change page title from `text-4xl font-bold` to `text-2xl font-semibold`
- Change subtitle from `text-[1.0625rem]` to `text-sm text-muted-foreground`
- Change section heading "Profile Information" from `text-xl font-semibold` to `text-lg font-semibold`
- Change field labels from `text-base font-medium` to `text-sm font-medium`
- Wrap profile sections in content cards instead of using dividers
- Normalize padding to `px-6 lg:px-12 py-10`

### 4. Installations page (`installations/page.tsx`)
- Change `max-w-7xl` to `max-w-6xl` (list/table tier)
- Normalize padding from `px-6 lg:px-12 py-8` to `px-6 lg:px-12 py-10`
- Page title already `text-2xl font-semibold` — no title change needed

### 5. Organization settings page (`installations/settings/organization/page.tsx`)
- Change section headings from `text-xl font-semibold` to `text-lg font-semibold`
- Change section subtitles from `text-base` to `text-sm text-muted-foreground`
- Standardize danger zone to use `text-destructive` and `border-destructive/20 bg-destructive/[0.03]`

### 6. Installation detail page (`installations/[id]/page.tsx`)
- Change danger zone heading from `text-xl font-semibold` to `text-lg font-semibold`
- Standardize danger zone cards to use `text-destructive` and `border-destructive/20 bg-destructive/[0.03]`
- Keep `AlertTriangle` icon in danger zone heading (consistent with warning semantics)

### 6b. Delete organization component (`components/delete-organization.tsx`)
- Normalize danger card from `border-destructive/30 bg-destructive/5` to `border-destructive/20 bg-destructive/[0.03]`

### 7. Installations list component (`components/installations-list.tsx`)
- Extract filter tabs into a new `FilterTabs` component that shares visual style with `NavTabs` but uses `<button>` + callback pattern

### 8. Billing page (`installations/settings/billing/page.tsx`)
- Change heading from `text-2xl font-bold` to `text-lg font-semibold`
- Add proper loading/empty/error states (not just a spinner)

### 9. Create installation shell (`components/installation-layout.tsx`)
- Change `max-w-7xl` to `max-w-4xl`
- Normalize padding from `px-6 lg:px-8 py-8` to `px-6 lg:px-12 py-10`

### 10. Installation settings tabs (`components/installation-settings-tabs.tsx`)
- Migrate to use shared `NavTabs` component to eliminate duplicate tab logic
- Normalize tab text from `text-base` to `text-sm font-medium`
- Normalize underline position from `bottom-1` to `bottom-0`

## Out of Scope

- Login page design (standalone, different context)
- Color scheme changes
- Component library (`@kloudlite/ui`) changes
- Mobile responsive improvements
