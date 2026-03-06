# Remaining React Best Practices Implementation Tasks

## Overview

All remaining tasks from the React Best Practices review that need to be completed. These tasks improve code quality, maintainability, and user experience.

---

## Medium Priority Tasks

### Task #6: Add ARIA live regions for dynamic content

**Status:** ✅ Completed

**Description:**
Add ARIA live regions to components that update dynamically to improve screen reader announcements. Key areas to address:

**Console App:**
- `installations-list.tsx` - Status filters, search results
- `team-members-table.tsx` - Updates after member management
- `team-invitations-table.tsx` - Updates after invitation actions

**Dashboard App:**
- `services-list.tsx` - Status updates
- `environments-list.tsx` - Status changes
- `packages-list.tsx` - Package list updates

**Website App:**
- `page.tsx` (homepage) - Page transitions, testimonial carousel
- Feature lists (multiple pages)

**Implementation Approach:**
- Add `aria-live="polite"` for important updates (status changes, errors)
- Add `aria-live="off"` for static content
- Use `aria-live="assertive"` for complete replacements when appropriate
- Ensure proper focus management for screen readers

**Files to modify:**
- Console: `/Users/karthik/dev/kloudlite-v2/web/apps/console/src/components/installations-list.tsx`
- `/Users/karthik/dev/kloudlite-v2/web/apps/console/src/components/team-members-table.tsx`
- `/Users/karthik/dev/kloudlite-v2/web/apps/console/src/components/team-invitations-table.tsx`
- Dashboard: `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/(main)/environments/_components/services-list.tsx`
- `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/(main)/environments/_components/environments-list.tsx`
- `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/(main)/workspaces/_components/packages-list.tsx`
- Website: `/Users/karthik/dev/kloudlite-v2/web/apps/website/src/app/(main)/page.tsx`

---

### Task #7: Add type guards to @kloudlite/types

**Status:** ✅ Completed

**Description:**
Add runtime type guard functions (isWorkspace, isEnvironment, isPackageRequest, etc.) to @kloudlite/types for better type safety and safer runtime checks

**Implementation Approach:**
- Create a `guards.ts` file in `/Users/karthik/dev/kloudlite-v2/web/packages/types/src/`
- Export type guard functions
- Add comprehensive JSDoc documentation
- Use TypeScript's type predicates for type narrowing

**Example signature:**
```typescript
export function isWorkspace(value: unknown): value is Workspace {
  return typeof value === 'object' && value !== null && 'apiVersion' in value && 'kind' in value
}
```

**Benefits:**
- Better type safety at runtime
- Clear error messages when type checks fail
- Easier to add new type guards
- Can be used throughout codebase

---

### Task #8: Consolidate duplicate utilities

**Status:** ✅ Completed

**Description:**
Consolidate duplicate utility functions across the codebase, particularly between @kloudlite/lib and website apps.

**Duplicates to address:**
- `cn()` function - exists in @kloudlite/lib and website
- `formatResourceName()` function - similar format utilities
- `formatWorkspaceName()` function - workspace name formatting
- Other string formatting utilities

**Files to modify:**
- Website: `/Users/karthik/dev/kloudlite-v2/web/apps/website/src/lib/utils.ts` (likely location)
- Search for other occurrences in website app

**Implementation Approach:**
1. Consolidate `cn()` into @kloudlite/lib
2. Remove duplicate functions from website
3. Import from @kloudlite/lib in website components
4. Update all imports throughout website app

**Benefits:**
- Single source of truth for utility
- Consistent behavior across codebase
- Reduced bundle size

---

### Task #9: Extract website page components

**Status:** ✅ Completed

**Description:**
Extract large components from website app `page.tsx` (621 lines) into smaller, testable pieces

**Large components in page.tsx:**
- TypewriterText component (animation effect)
- WorkflowVisualization component (interactive workflow)
- FeatureCard component (feature cards)
- Pricing section (pricing cards, tiers)
- Testimonials section (carousel)
- Hero section (content and CTA)
- Navbar/Header components

**Implementation Approach:**
- Create new directory: `/Users/karthik/dev/kloudlite-v2/web/apps/website/src/components/home-page/`
- Extract each component to its own file
- Maintain proper exports for existing components
- Add proper TypeScript interfaces

---

## Task #4 Status: ✅ Completed (Rechecked + Implemented)

Task #4 (split dashboard `workspace.actions.ts`) is now implemented using a safe Next.js server action pattern:

- Split monolith into direct-import action modules (no barrel re-exports):
  - `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/actions/workspace-query.actions.ts`
  - `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/actions/workspace-mutation.actions.ts`
  - `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/actions/workspace-packages.actions.ts`
  - `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/actions/workspace-code-analysis.actions.ts`
- Extracted shared helper:
  - `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/actions/workspace.actions.shared.ts`
- Extracted code-analysis types:
  - `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/actions/workspace-code-analysis.types.ts`
- Updated all dashboard imports to point to specific action files.
- Removed old monolith:
  - `/Users/karthik/dev/kloudlite-v2/web/apps/dashboard/src/app/actions/workspace.actions.ts`

---

## Completed Tasks (7 tasks)

### Foundation Phase ✅
1. **Setup test infrastructure for shared packages** - Added vitest, React Testing Library, happy-dom to both @kloudlite/ui and @kloudlite/lib
2. **Add ErrorBoundary component to @kloudlite/ui** - Reusable error boundary with fallback UI
3. **Fix window.location.reload() usage** - Replaced with `router.refresh()`
4. **Fix window.location.href usage** - Replaced with `router.push()`

### High Priority Phase ✅
5. **Add useCallback to inline event handlers** - Fixed inline arrow functions in website-header, installations-list, and team-members-table
6. **Add React.memo to large list components** - Added memoization to 17 components (later refined to 3 stateless components)
7. **Add key props to large list components** - Fixed array index keys in 16 files

## Implementation Notes

- All test infrastructure is set up and ready for test writing
- ErrorBoundary component is available and exported
- React Router navigation patterns are fixed throughout console app
- Components are properly memoized where beneficial
- Key props now use unique identifiers throughout web codebase

## Getting Started

To work on a task, simply tell me: "I'll work on Task #X" and I'll dispatch an implementer subagent.

All task context, research, and file information from the original analysis has been preserved in the auto-memory system.

**Next available task:** Task #7 (Add type guards to @kloudlite/types) - Good starting point for backend-type work
