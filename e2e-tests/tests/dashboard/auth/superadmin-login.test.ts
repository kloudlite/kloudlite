import { test, expect } from '@playwright/test'
import { dashboardTest } from '../../../lib/fixtures'
import { DEV_TOKEN } from '../../../lib/constants'

test.describe('Dashboard > Auth > Super Admin Login', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('login with dev token and redirect to admin', async ({ page }) => {
    await page.goto(`/superadmin-login?token=${DEV_TOKEN}`)

    // Server-side signIn redirects straight to /admin — no intermediate page shown
    await page.waitForURL('**/admin/**', { timeout: 15_000 })
    await expect(page.getByRole('banner').getByText('Admin', { exact: true })).toBeVisible()
  })

  test('missing token shows error', async ({ page }) => {
    await page.goto('/superadmin-login')

    await expect(page.getByText('Missing authentication token')).toBeVisible()
    await expect(page.getByText('Authentication failed')).toBeVisible()
  })

  test('invalid token shows error', async ({ page }) => {
    await page.goto('/superadmin-login?token=bad-token-value')

    await expect(page.getByText('Authentication failed')).toBeVisible({ timeout: 15_000 })
  })
})

dashboardTest.describe('Dashboard > Auth > Super Admin Dashboard', () => {
  dashboardTest.use({ storageState: { cookies: [], origins: [] } })

  dashboardTest('nav links are visible and navigable', async ({ page }) => {
    await page.getByRole('link', { name: 'Machine Configs' }).click()
    await expect(page).toHaveURL(/\/admin\/machine-configs/)

    await page.getByRole('link', { name: 'OAuth Providers' }).click()
    await expect(page).toHaveURL(/\/admin\/oauth-providers/)

    await page.getByRole('link', { name: 'User Management' }).click()
    await expect(page).toHaveURL(/\/admin\/users/)
  })

  dashboardTest('sign out redirects away from admin', async ({ page }) => {
    await page.getByRole('button', { name: 'Super Admin (Dev)' }).click()

    await expect(page.getByText('admin@dev-installation')).toBeVisible()

    await page.getByRole('menuitem', { name: 'Sign out' }).click()
    await expect(page).not.toHaveURL(/\/admin/)
  })
})
