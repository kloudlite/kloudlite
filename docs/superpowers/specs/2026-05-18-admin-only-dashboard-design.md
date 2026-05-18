# Admin-only dashboard design

## Goal

Convert `web/apps/dashboard` into an administrative dashboard. User-facing workspace, environment, artifact, code-analysis, and personal dashboard pages move out of scope for this web app and will be rebuilt in the desktop app when needed.

The dashboard should make admin the primary experience and use the current `platform-api` server as its backend API target.

## Non-goals

- Do not preserve user-facing routes as hidden or redirect-only pages beyond the root redirect.
- Do not rebuild desktop user functionality in the web dashboard.
- Do not redesign admin UI styling unless required by route/API cleanup.
- Do not change backend `platform-api` behavior as part of this dashboard cleanup unless a missing admin endpoint blocks the retained admin pages.

## Retained routes

Keep the administrative and auth entry points:

- `/` redirects to `/admin`.
- `/admin` remains the primary authenticated app shell.
- `/admin/users` remains for user administration.
- `/admin/machine-configs` remains for machine type/work machine administration where applicable.
- `/admin/oauth-providers` remains for OAuth provider configuration.
- `/auth/signin`, `/auth/error`, and `/superadmin-login` remain.
- Existing API routes required by retained auth/admin functionality remain.

## Removed routes

Delete the user-facing route group under `src/app/(main)`:

- dashboard landing pages
- workspaces and workspace detail pages
- environments, environment services, configs, and settings pages
- artifacts pages
- code-analysis pages

Delete user-only navigation/chrome components when they become unreferenced. Keep shared visual primitives and admin components that are still referenced by retained pages.

## Access behavior

Admin becomes the primary application mode:

- Authenticated admins and super-admins land in `/admin`.
- Super-admin token login continues to land in admin.
- Non-admin users should not enter the dashboard app. If they reach `/`, they are redirected through the admin auth guard and then away according to existing auth behavior.
- The admin layout continues to require `admin` or `super-admin` roles.

## API server integration

The dashboard API client should target the current `platform-api` server.

- Development fallback should point at `https://localhost:9443`, not the old `http://localhost:8080` default.
- `NEXT_PUBLIC_API_URL` remains the configurable API base URL.
- Admin data services should use the current platform resource API shape where possible.
- Keep the dashboard-side auth token behavior: requests attach the JWT from the current session/super-admin token.

Admin pages that already use `apiClient` should continue through that abstraction. If an admin service currently assumes an obsolete API route, update that service rather than embedding fetch logic in pages.

## Component/data boundaries

- Route files own page composition only.
- Admin pages should depend on admin components and service modules.
- `apiClient` remains the HTTP boundary for platform-api calls.
- Shared utilities/types can remain if still imported by retained admin/auth/API code.
- User-only services, validations, hooks, and components should be removed only when no retained code imports them.

## Error handling

- Missing/invalid session keeps redirecting to `/auth/signin`.
- Authenticated non-admin access to `/admin` remains blocked by the admin layout.
- Failed admin data loads should keep existing admin error states where present.
- API client errors should continue surfacing backend error messages when available.

## Testing and verification

Run focused dashboard verification from `web/apps/dashboard`:

```bash
bun run test:run
bun run lint
```

If lint/test coverage is incomplete for deleted routes, also run a production build or TypeScript check through the package scripts available in `package.json`.

Manual verification target:

- `/` redirects to `/admin`.
- `/admin/users`, `/admin/machine-configs`, and `/admin/oauth-providers` render or fail with existing admin error states, not missing user-route imports.
- Removed user URLs no longer have page implementations.

## Implementation notes

Start with route deletion and root redirect, then fix imports and services revealed by TypeScript/tests. Avoid broad UI rewrites. Keep commits scoped and reviewable.
