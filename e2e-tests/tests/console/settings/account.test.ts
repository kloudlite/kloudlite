import { consoleTest as test, expect } from '../../../lib/fixtures'

test.describe('Console > Settings > Account', () => {
  test.use({ storageState: { cookies: [], origins: [] } })

  test('navigate to account settings via header dropdown', async ({ page }) => {
    // Wait for the page to fully load before interacting with header
    await expect(page.getByRole('heading', { name: 'Installations' })).toBeVisible()

    await page.getByRole('button', { name: /Karthik/ }).click()
    await page.getByRole('menuitem', { name: 'Account Settings' }).click()

    await expect(page).toHaveURL(/\/installations\/settings/, { timeout: 15_000 })
    await expect(page.getByRole('heading', { name: 'Account Settings' })).toBeVisible()
    await expect(page.getByText('Manage your account information and preferences')).toBeVisible()
  })

  test('profile tab shows user information', async ({ page }) => {
    await page.goto('/installations/settings/profile')

    await expect(page.getByRole('heading', { name: 'Profile Information' })).toBeVisible({ timeout: 15_000 })
    // Use the profile content area (not header) to find user details
    const profileSection = page.locator('main')
    await expect(profileSection.getByText('Karthik', { exact: true })).toBeVisible()
    await expect(profileSection.getByText('karthik@kloudlite.io')).toBeVisible()
  })

  test('profile page shows authentication provider', async ({ page }) => {
    await page.goto('/installations/settings/profile')

    await expect(page.getByText('Authentication Provider')).toBeVisible({ timeout: 15_000 })
  })

  test('profile page shows name and email labels', async ({ page }) => {
    await page.goto('/installations/settings/profile')

    await expect(page.getByRole('heading', { name: 'Profile Information' })).toBeVisible({ timeout: 15_000 })
    const profileSection = page.locator('main')
    await expect(profileSection.getByText('Name', { exact: true })).toBeVisible()
    await expect(profileSection.getByText('Email Address')).toBeVisible()
  })

  test('settings page has profile and billing tabs', async ({ page }) => {
    await page.goto('/installations/settings/profile')

    await expect(page.getByRole('link', { name: 'Profile' })).toBeVisible({ timeout: 15_000 })
    await expect(page.getByRole('link', { name: 'Billing' })).toBeVisible()
  })

  test('billing tab is navigable', async ({ page }) => {
    await page.goto('/installations/settings/profile')

    await expect(page.getByRole('link', { name: 'Billing' })).toBeVisible({ timeout: 15_000 })
    await page.getByRole('link', { name: 'Billing' }).click()
    await expect(page).toHaveURL(/\/installations\/settings\/billing/)
  })
})
