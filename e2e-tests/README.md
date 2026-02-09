# E2E Tests

End-to-end tests for the Kloudlite platform using [Playwright](https://playwright.dev).

## Prerequisites

- **Bun** runtime installed
- **Dashboard** running on `localhost:3001`
- **Console** running on `localhost:3002`

```bash
# Install dependencies (first time)
bun install

# Install Playwright browsers (first time)
bunx playwright install
```

## Running Tests

```bash
# Run all UI tests (dashboard + console, excludes provider tests)
bun run test

# Run by project
bun run test:dashboard      # Dashboard tests only
bun run test:console        # Console tests only (excludes providers)

# Run provider lifecycle tests (long-running, requires cloud credentials)
bun run test:aws
bun run test:gcp
bun run test:azure

# Interactive mode
bun run test:ui             # Playwright UI
bun run test:headed         # Visible browser

# Run a single test file
bunx playwright test tests/dashboard/auth/superadmin-login.test.ts
```

## Project Structure

```
e2e-tests/
  lib/                          # Shared test infrastructure
    constants.ts                # Centralized tokens, timeouts
    helpers.ts                  # Reusable functions (login, shell spawn, etc.)
    fixtures.ts                 # Custom Playwright fixtures (pre-authenticated pages)
    provider-test.ts            # Shared provider installation lifecycle

  tests/
    console/                    # Console app tests (localhost:3002)
      global.setup.ts           # Warm-up: pre-compiles Turbopack chunks
      auth/
        login.test.ts           # Login page UI, OAuth buttons, dev login flow, sign out
      installations/
        list.test.ts            # Installation list, filters, search, table columns
        detail.test.ts          # Installation detail, overview/team tabs
        cli-download.test.ts    # CLI download page, platform sections, direct links
      providers/
        aws/installation.test.ts    # AWS installation lifecycle + form validation
        gcp/installation.test.ts    # GCP installation lifecycle
        azure/installation.test.ts  # Azure installation lifecycle
      settings/
        account.test.ts         # Account settings, profile info, billing tab
        theme.test.ts           # Dark/light/system theme switching, persistence

    dashboard/                  # Dashboard app tests (localhost:3001)
      global.setup.ts           # Warm-up: pre-compiles Turbopack chunks
      auth/
        superadmin-login.test.ts    # Superadmin login flow, error states, post-login nav
        oauth-providers.test.ts     # OAuth provider config, toggle state, login screen
        existing-session.test.ts    # Active session warning when switching accounts
      users/
        management.test.ts      # Add/edit/delete users, form validation, role selection
        role-access.test.ts     # Role-based access control (user-only, admin-only, both)
      settings/
        machine-configs.test.ts # Machine configuration form, validation, categories
        theme.test.ts           # Dark/light/system theme switching, persistence
```

## Architecture

### Shared Library (`lib/`)

All shared code lives in `lib/` to eliminate duplication across test files.

#### `constants.ts`

Centralized values used across tests:

| Constant | Value | Purpose |
|----------|-------|---------|
| `DEV_TOKEN` | `'dev-superadmin'` | Development superadmin authentication token |
| `TIMEOUTS.action` | 10s | UI action timeout (clicks, fills) |
| `TIMEOUTS.navigation` | 15s | Page navigation timeout |
| `TIMEOUTS.auth` | 15s | Authentication redirect timeout |
| `TIMEOUTS.deployment` | 10min | Wait for cloud deployment detection |
| `TIMEOUTS.dns` | 15min | Wait for DNS propagation + LB health |
| `TIMEOUTS.providerTest` | 30min | Overall provider test timeout |

#### `helpers.ts`

Reusable functions:

| Function | Purpose |
|----------|---------|
| `devLogin(page)` | Console dev login via `/api/dev-login`, waits for redirect to `/installations` |
| `superadminLogin(page, token?)` | Dashboard superadmin login via `/superadmin-login?token=...`, waits for `/admin` |
| `runScript(command, cwd, label)` | Spawns a shell command with streaming stdout/stderr, returns `{ process, done }` |
| `escapeRegex(str)` | Escapes special regex characters in a string |

#### `fixtures.ts`

Custom [Playwright test fixtures](https://playwright.dev/docs/test-fixtures) that provide pre-authenticated pages:

| Fixture | Auth Method | Landing Page |
|---------|-------------|--------------|
| `consoleTest` | `devLogin()` | `/installations` |
| `dashboardTest` | `superadminLogin()` | `/admin/**` |

Usage in test files:

```typescript
// Console tests — page arrives at /installations, already logged in
import { consoleTest as test, expect } from '../../../lib/fixtures'

test('page heading visible', async ({ page }) => {
  await expect(page.getByRole('heading', { name: 'Installations' })).toBeVisible()
})
```

```typescript
// Dashboard tests — page arrives at /admin, already logged in
import { dashboardTest as test, expect } from '../../../lib/fixtures'

test('user management loads', async ({ page }) => {
  await page.getByRole('link', { name: 'User Management' }).click()
  // ...
})
```

Tests that verify the login flow itself (e.g., `login.test.ts`, `superadmin-login.test.ts`) use the raw `test` import from `@playwright/test` instead.

#### `provider-test.ts`

The three provider tests (AWS, GCP, Azure) share a near-identical lifecycle. `runProviderInstallationTest()` encapsulates the full flow:

1. Login and create a new installation
2. Fill form (name, subdomain) and submit
3. Switch to provider tab, verify prerequisites and install command
4. Execute install script via `runScript()` (spawns CLI process)
5. Wait for deployment detection in UI
6. Wait for DNS propagation and redirect to complete page
7. Verify installation is active, URL is reachable
8. Navigate to settings, verify status and superadmin login URL
9. Run uninstall script to tear down cloud resources
10. Delete installation record, verify removal

Each provider test file configures the shared lifecycle with provider-specific values:

```typescript
import { consoleTest } from '../../../../lib/fixtures'
import { runProviderInstallationTest } from '../../../../lib/provider-test'

consoleTest.describe('Console > Providers > GCP > Installation', () => {
  consoleTest.use({ storageState: { cookies: [], origins: [] } })

  runProviderInstallationTest(consoleTest, {
    name: 'GCP',
    tabName: 'GCP',
    prerequisiteText: 'gcloud CLI configured with Application Default Credentials',
    regionLabel: 'Select GCP Region:',
    installUrlPath: 'install/gcp',
    uninstallUrlPath: 'uninstall/gcp',
    subdomainPrefix: 'e2e-gcp-',
    testNamePrefix: 'E2E GCP',
    tmpDirPrefix: 'kl-e2e-gcp-',
  })
})
```

### Playwright Configuration

Defined in `playwright.config.ts`. Six projects:

| Project | Directory | Depends On | Timeout | Purpose |
|---------|-----------|------------|---------|---------|
| `warmup-dashboard` | `tests/dashboard` | - | 30s | Pre-compile Turbopack chunks |
| `dashboard` | `tests/dashboard` | warmup-dashboard | 30s | Dashboard UI tests |
| `warmup-console` | `tests/console` | - | 30s | Pre-compile Turbopack chunks |
| `console` | `tests/console` | warmup-console | 30s | Console UI tests (excludes `providers/`) |
| `aws` | `tests/console/providers/aws` | warmup-console | 30min | AWS installation lifecycle |
| `gcp` | `tests/console/providers/gcp` | warmup-console | 30min | GCP installation lifecycle |
| `azure` | `tests/console/providers/azure` | warmup-console | 30min | Azure installation lifecycle |

The `console` project explicitly ignores `providers/` so provider tests only run when targeted by project name.

## Writing New Tests

1. **Console tests requiring auth**: Import `consoleTest as test` from `lib/fixtures`. The page arrives pre-authenticated at `/installations`.

2. **Dashboard tests requiring auth**: Import `dashboardTest as test` from `lib/fixtures`. The page arrives pre-authenticated at `/admin`.

3. **Tests verifying login flows**: Import raw `test` from `@playwright/test` and handle auth manually.

4. **Timeouts**: Import from `lib/constants` instead of using magic numbers.

5. **Storage state**: Always set `test.use({ storageState: { cookies: [], origins: [] } })` to ensure a clean session per test.
