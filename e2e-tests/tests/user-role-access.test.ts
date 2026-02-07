import { test, expect, Page } from '@playwright/test'

const DEV_TOKEN = 'dev-superadmin'

// Test user credentials
const TEST_PASSWORD = 'testpass123'
const TEST_USERS = {
  userOnly: {
    email: 'test-user-only@e2e.local',
    username: 'test-user-only',
    roles: ['User'],
  },
  adminOnly: {
    email: 'test-admin-only@e2e.local',
    username: 'test-admin-only',
    roles: ['Admin'],
  },
  bothRoles: {
    email: 'test-both-roles@e2e.local',
    username: 'test-both-roles',
    roles: ['User', 'Admin'],
  },
} as const

// --- Helpers ---

async function loginAsSuperAdmin(page: Page) {
  await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)
  await page.waitForURL('**/admin/**', { timeout: 15_000 })
}

async function navigateToUserManagement(page: Page) {
  await page.getByRole('link', { name: 'User Management' }).click()
  await expect(
    page.getByRole('heading', { level: 1, name: 'User Management' }),
  ).toBeVisible()
}

async function createUser(
  page: Page,
  user: { email: string; roles: readonly string[] },
) {
  // Skip if user already exists in the table
  const existingRow = page.getByRole('row').filter({ hasText: user.email })
  if ((await existingRow.count()) > 0) return

  await page.getByRole('button', { name: 'Add User' }).click()
  const dialog = page.getByRole('dialog')
  await expect(dialog).toBeVisible()

  // Fill email (username auto-fills)
  await dialog.getByPlaceholder('Enter email address').fill(user.email)

  // Wait for username to auto-populate
  await expect(
    dialog.getByPlaceholder('Enter username (e.g., john-doe)'),
  ).not.toHaveValue('')

  // Select roles
  for (const role of user.roles) {
    await dialog.getByRole('button', { name: role, exact: true }).click()
  }

  // Submit
  await dialog.getByRole('button', { name: 'Create' }).click()

  // Wait for dialog to close (indicates success)
  await expect(dialog).toBeHidden({ timeout: 15_000 })
}

async function setUserPassword(page: Page, email: string, password: string) {
  // Find the row containing this user's email and click its actions menu
  const row = page.getByRole('row').filter({ hasText: email })
  await row.getByRole('button').click()

  await page.getByRole('menuitem', { name: 'Reset Password' }).click()

  const dialog = page.getByRole('dialog', { name: 'Reset Password' })
  await expect(dialog).toBeVisible()

  await page
    .getByPlaceholder('Enter new password (min 8 characters)')
    .fill(password)
  await dialog.getByRole('button', { name: 'Reset Password' }).click()

  // Wait for dialog to close
  await expect(dialog).toBeHidden({ timeout: 10_000 })
}

async function deleteUser(page: Page, email: string) {
  const row = page.getByRole('row').filter({ hasText: email })
  // Check if the row exists before attempting delete
  if ((await row.count()) === 0) return

  await row.getByRole('button').click()
  await page.getByRole('menuitem', { name: 'Delete User' }).click()

  const dialog = page.getByRole('dialog', { name: 'Delete User' })
  await expect(dialog).toBeVisible()
  await dialog.getByRole('button', { name: 'Delete User' }).click()

  // Wait for dialog to close
  await expect(dialog).toBeHidden({ timeout: 10_000 })
}

async function loginAsCredentials(page: Page, email: string, password: string) {
  await page.goto('/auth/signin')
  await page.getByRole('textbox', { name: 'Email address' }).fill(email)
  await page.getByRole('textbox', { name: 'Password' }).fill(password)
  await page.getByRole('button', { name: 'Sign in' }).click()
  await page.waitForURL(
    (url) => !url.pathname.includes('/auth/signin'),
    { timeout: 30_000 },
  )
}

test.describe('Role-based user access', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  // Increase timeout for setup/teardown since they create multiple users
  test.describe.configure({ timeout: 120_000 })

  // --- Setup: create test users as superadmin ---
  test.beforeAll(async ({ browser }) => {
    const page = await browser.newPage({ storageState: { cookies: [], origins: [] } })

    await loginAsSuperAdmin(page)
    await navigateToUserManagement(page)

    // Create all three test users
    for (const user of Object.values(TEST_USERS)) {
      await createUser(page, user)
    }

    // Set passwords for each user
    for (const user of Object.values(TEST_USERS)) {
      await setUserPassword(page, user.email, TEST_PASSWORD)
    }

    await page.close()
  })

  // --- Teardown: delete test users ---
  test.afterAll(async ({ browser }) => {
    const page = await browser.newPage({ storageState: { cookies: [], origins: [] } })

    await loginAsSuperAdmin(page)
    await navigateToUserManagement(page)

    // Delete all test users (reverse order to avoid index shifts)
    for (const user of Object.values(TEST_USERS).reverse()) {
      await deleteUser(page, user.email)
    }

    await page.close()
  })

  // --- Group 1: User-only role ---
  test.describe('User-only role', () => {
    test('can access main routes but not admin', async ({ page }) => {
      await loginAsCredentials(page, TEST_USERS.userOnly.email, TEST_PASSWORD)

      // Should land on a main route (not /admin)
      expect(page.url()).not.toContain('/admin')

      // Navigate to /admin → should redirect back to main
      await page.goto('/admin')
      await page.waitForURL((url) => !url.pathname.startsWith('/admin'), {
        timeout: 15_000,
      })
      expect(page.url()).not.toContain('/admin')
    })

    test('profile dropdown does not show Administration link', async ({ page }) => {
      await loginAsCredentials(page, TEST_USERS.userOnly.email, TEST_PASSWORD)

      // Open profile dropdown — button contains the username
      await page
        .getByRole('button', { name: new RegExp(TEST_USERS.userOnly.username) })
        .click()

      // "Administration" link should NOT be present (user-only has no admin role)
      await expect(
        page.getByRole('menuitem', { name: 'Administration' }),
      ).toBeHidden()

      // "Sign Out" should be present
      await expect(
        page.getByRole('menuitem', { name: 'Sign Out' }),
      ).toBeVisible()
    })
  })

  // --- Group 2: Admin-only role ---
  test.describe('Admin-only role', () => {
    test('redirected to admin and cannot access main routes', async ({ page }) => {
      await loginAsCredentials(page, TEST_USERS.adminOnly.email, TEST_PASSWORD)

      // Should be redirected to /admin (admin-only users get redirected from main)
      await page.waitForURL('**/admin/**', { timeout: 15_000 })
      expect(page.url()).toContain('/admin')

      // Can access admin users page
      await page.getByRole('link', { name: 'User Management' }).click()
      await expect(
        page.getByRole('heading', { level: 1, name: 'User Management' }),
      ).toBeVisible()

      // Navigate to / → should redirect back to /admin
      await page.goto('/')
      await page.waitForURL('**/admin/**', { timeout: 15_000 })
      expect(page.url()).toContain('/admin')
    })

    test('admin profile dropdown does not show Dashboard link', async ({ page }) => {
      await loginAsCredentials(page, TEST_USERS.adminOnly.email, TEST_PASSWORD)
      await page.waitForURL('**/admin/**', { timeout: 15_000 })

      // Open admin profile dropdown — button contains the username
      await page
        .getByRole('button', { name: new RegExp(TEST_USERS.adminOnly.username) })
        .click()

      // "Dashboard" link should NOT be present (no user role)
      await expect(
        page.getByRole('menuitem', { name: 'Dashboard' }),
      ).toBeHidden()

      // "Sign out" should be present (note: lowercase 'o' in admin dropdown)
      await expect(
        page.getByRole('menuitem', { name: 'Sign out' }),
      ).toBeVisible()
    })
  })

  // --- Group 3: Both roles (user + admin) ---
  test.describe('Both roles (user + admin)', () => {
    test('can access both main and admin, and switch between them', async ({ page }) => {
      await loginAsCredentials(page, TEST_USERS.bothRoles.email, TEST_PASSWORD)

      // Should land on main route (has user role)
      expect(page.url()).not.toContain('/admin')

      // Open profile dropdown → "Administration" link should be present
      await page
        .getByRole('button', { name: new RegExp(TEST_USERS.bothRoles.username) })
        .click()
      await expect(
        page.getByRole('menuitem', { name: 'Administration' }),
      ).toBeVisible()

      // Click "Administration" → navigate to /admin
      await page.getByRole('menuitem', { name: 'Administration' }).click()
      await page.waitForURL('**/admin/**', { timeout: 15_000 })
      expect(page.url()).toContain('/admin')

      // Open admin profile dropdown → "Dashboard" link should be present
      await page
        .getByRole('button', { name: new RegExp(TEST_USERS.bothRoles.username) })
        .click()
      await expect(
        page.getByRole('menuitem', { name: 'Dashboard' }),
      ).toBeVisible()

      // Click "Dashboard" → navigate back to main
      await page.getByRole('menuitem', { name: 'Dashboard' }).click()
      await page.waitForURL(
        (url) => !url.pathname.startsWith('/admin'),
        { timeout: 15_000 },
      )
      expect(page.url()).not.toContain('/admin')

      // Open the main app profile dropdown and verify "Administration" link is present
      await page
        .getByRole('button', { name: new RegExp(TEST_USERS.bothRoles.username) })
        .click()
      await expect(
        page.getByRole('menuitem', { name: 'Administration' }),
      ).toBeVisible()
    })
  })
})
