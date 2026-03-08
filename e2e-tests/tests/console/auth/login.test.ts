import { test, expect } from '@playwright/test'
import { devLogin } from '../../../lib/helpers'

test.use({ storageState: { cookies: [], origins: [] } })

test.describe('Console > Auth', () => {
  test('login page shows OAuth providers', async ({ page }) => {
    await page.goto('/login')

    await expect(
      page.getByRole('heading', { name: /welcome/i }),
    ).toBeVisible()
    await expect(
      page.getByRole('link', { name: /continue with github/i }),
    ).toBeVisible()
    await expect(
      page.getByRole('link', { name: /continue with google/i }),
    ).toBeVisible()
  })

  test('dev login redirects to installations page', async ({ page }) => {
    await devLogin(page)

    await expect(page).toHaveURL(/\/installations$/)
    await expect(
      page.getByRole('heading', { name: 'Installations' }),
    ).toBeVisible()
  })

  test('org switcher is visible after login', async ({ page }) => {
    await devLogin(page)

    // The org switcher should be visible in the nav bar
    // It shows the org name (e.g., "Kloudlite Company")
    const orgSwitcher = page.getByRole('button', { name: /company/i })
    await expect(orgSwitcher).toBeVisible()
  })
})
